// Package session handles IMAP commands received from clients
// within a single IMAP session (one client connection).
package session

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/gluon/events"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/imap/command"
	"github.com/ProtonMail/gluon/internal/backend"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/ProtonMail/gluon/limits"
	"github.com/ProtonMail/gluon/profiling"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/gluon/rfcparser"
	"github.com/ProtonMail/gluon/version"
	"github.com/emersion/go-imap/utf7"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

const maxSessionError = 20

type Session struct {
	// conn is the underlying TCP connection to the client. It is wrapped by a buffered liner.
	conn net.Conn

	scanner        *rfcparser.Scanner
	inputCollector *command.InputCollector

	// backend provides access to the IMAP backend (including the database).
	backend *backend.Backend

	// state manages state of the authorized backend for this session.
	state *state.State

	// userLock protects the session's user object.
	userLock sync.Mutex

	// caps is the server's IMAP caps.
	caps []imap.Capability

	// capsLock protects the session's capabilities object.
	capsLock sync.Mutex

	// sessionID is this session's unique ID.
	sessionID int

	// eventCh is a channel on which the session should publish any events that occur.
	eventCh chan<- events.Event

	// loggers which can be set to log incoming and outgoing IMAP communications.
	inLogger, outLogger io.Writer

	// tlsConfig holds TLS information (used, for example, for STARTTLS).
	tlsConfig *tls.Config

	// idleBulkTime to control how often IDLE responses are sent. 0 means
	// immediate response with no response merging.
	idleBulkTime time.Duration

	// imapID holds the IMAP ID extension data for this client. This is necessary, since this information may arrive
	// before the client logs in or selects a mailbox.
	imapID imap.IMAPID

	// version is the version info of the Gluon server.
	version version.Info

	// cmdProfilerBuilder is used in profiling command execution.
	cmdProfilerBuilder profiling.CmdProfilerBuilder

	// handleWG is used to wait for all commands to finish before closing the session.
	handleWG async.WaitGroup

	/// errorCount error counter
	errorCount int

	imapLimits limits.IMAP

	panicHandler async.PanicHandler

	log *logrus.Entry
}

func New(
	conn net.Conn,
	backend *backend.Backend,
	sessionID int,
	version version.Info,
	profiler profiling.CmdProfilerBuilder,
	eventCh chan<- events.Event,
	idleBulkTime time.Duration,
	panicHandler async.PanicHandler,
) *Session {
	inputCollector := command.NewInputCollector(bufio.NewReader(conn))
	scanner := rfcparser.NewScannerWithReader(inputCollector)

	return &Session{
		conn:               conn,
		inputCollector:     inputCollector,
		scanner:            scanner,
		backend:            backend,
		caps:               []imap.Capability{imap.IMAP4rev1, imap.IDLE, imap.UNSELECT, imap.UIDPLUS, imap.MOVE, imap.ID},
		sessionID:          sessionID,
		eventCh:            eventCh,
		idleBulkTime:       idleBulkTime,
		version:            version,
		cmdProfilerBuilder: profiler,
		handleWG:           async.MakeWaitGroup(panicHandler),
		panicHandler:       panicHandler,
		log:                logrus.WithField("pkg", "gluon/session").WithField("session", sessionID),
	}
}

func (s *Session) SetIncomingLogger(w io.Writer) {
	if w == nil {
		panic("setting a nil writer")
	}

	s.inLogger = w
}

func (s *Session) SetOutgoingLogger(w io.Writer) {
	if w == nil {
		panic("setting a nil writer")
	}

	s.outLogger = w
}

func (s *Session) SetTLSConfig(cfg *tls.Config) {
	if cfg == nil {
		panic("setting a nil TLS config")
	}

	s.tlsConfig = cfg

	s.addCapability(imap.StartTLS)
}

func (s *Session) Serve(ctx context.Context) error {
	defer s.done(ctx)
	defer s.handleWG.Wait()

	if err := s.greet(); err != nil {
		return err
	}

	profiler := s.cmdProfilerBuilder.New()
	defer s.cmdProfilerBuilder.Collect(profiler)

	return s.serve(profiling.WithProfiler(ctx, profiler))
}

func (s *Session) serve(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cmdCh := s.startCommandReader(ctx)

	for {
		select {
		case update := <-s.state.GetStateUpdatesCh():
			if err := s.state.ApplyUpdate(ctx, update); err != nil {
				s.log.WithError(err).Error("Failed to apply state update")
			}

			continue

		case res, ok := <-cmdCh:
			if !ok {
				return nil
			}

			if res.err != nil {
				if err := response.Bad(res.command.Tag).WithError(res.err).Send(s); err != nil {
					return err
				}

				if s.errorCount += 1; s.errorCount >= maxSessionError {
					// there's no events like this in sentry so far.
					reporter.MessageWithContext(ctx,
						"Too many errors in session, closing connection",
						reporter.Context{"ID": s.imapID.String()},
					)

					return nil
				}

				continue
			} else {
				s.errorCount = 0
			}

			// Before proceeding with command execution, check whether we still have a valid state.
			// State can become invalid at any time, e.g.: deletion of a selected mailbox by another client.
			if s.state != nil && !s.state.IsValid() {
				return response.Bye().WithInconsistentState().Send(s)
			}

			switch cmd := res.command.Payload.(type) {
			case *command.Logout:
				return s.handleLogout(ctx, res.command.Tag, cmd)

			case *command.Idle:
				if err := s.handleIdle(ctx, res.command.Tag, cmd, cmdCh); err != nil {
					if err := response.No(res.command.Tag).WithError(err).Send(s); err != nil {
						return fmt.Errorf("failed to send response to client: %w", err)
					}
				}

			default:
				respCh := s.handleOther(withStartTime(ctx, time.Now()), res.command.Tag, cmd)
				for res := range respCh {
					if err := res.Send(s); err != nil {
						go func() {
							defer async.HandlePanic(s.panicHandler)

							for range respCh {
								// Consume all invalid input on error that is still being produced by the ongoing
								// command.
							}
						}()

						return fmt.Errorf("failed to send response to client: %w", err)
					}
				}
			}

		case <-s.state.Done():
			return nil

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *Session) WriteResponse(res string) error {
	s.logOutgoing(res)

	if _, err := s.conn.Write([]byte(res + "\r\n")); err != nil {
		return err
	}

	return nil
}

func (s *Session) logIncoming(line string) {
	if s.inLogger == nil {
		return
	}

	writeLog(s.inLogger, "C", strconv.Itoa(s.sessionID), line)
}

func (s *Session) logOutgoing(line string) {
	if s.outLogger == nil {
		return
	}

	writeLog(s.outLogger, "S", strconv.Itoa(s.sessionID), line)
}

func (s *Session) done(ctx context.Context) {
	close(s.eventCh)

	if s.state != nil {
		if err := s.state.ReleaseState(ctx); err != nil {
			s.log.WithError(err).Error("Failed to close state")
		}
	}

	_ = s.conn.Close()
}

func (s *Session) addCapability(capability imap.Capability) {
	s.capsLock.Lock()
	defer s.capsLock.Unlock()

	if !slices.Contains(s.caps, capability) {
		s.caps = append(s.caps, capability)
	}
}

func (s *Session) remCapability(capability imap.Capability) {
	s.capsLock.Lock()
	defer s.capsLock.Unlock()

	if idx := slices.Index(s.caps, capability); idx >= 0 {
		s.caps = slices.Delete(s.caps, idx, idx+1)
	}
}

func (s *Session) greet() error {
	s.capsLock.Lock()
	defer s.capsLock.Unlock()

	return response.Ok().
		WithItems(response.ItemCapability(s.caps...)).
		WithMessage(fmt.Sprintf("%v %v - gluon session ID %v", s.version.Name, s.version.Version.String(), s.sessionID)).
		Send(s)
}

func (s *Session) decodeMailboxName(name string) (string, error) {
	delimiter := s.backend.GetDelimiter()

	split := strings.SplitAfterN(name, delimiter, 2)
	if !strings.EqualFold(split[0], fmt.Sprintf("INBOX%v", delimiter)) || len(split) != 2 {
		return utf7.Encoding.NewDecoder().String(name)
	}

	return utf7.Encoding.NewDecoder().String(fmt.Sprintf("INBOX%v%v", delimiter, split[1]))
}
