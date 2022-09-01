// Package session handles IMAP commands received from clients
// within a single IMAP session (one client connection).
package session

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/ProtonMail/gluon/events"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal"
	"github.com/ProtonMail/gluon/internal/backend"
	"github.com/ProtonMail/gluon/internal/liner"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/ProtonMail/gluon/profiling"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

type Session struct {
	// conn is the underlying TCP connection to the client. It is wrapped by a buffered liner.
	conn net.Conn

	// liner wraps the underlying TCP connection to facilitate linewise reading.
	liner *liner.Liner

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

	// imapID holds the IMAP ID extension data for this client. This is necessary, since this information may arrive
	// before the client logs in or selects a mailbox.
	imapID imap.ID

	version *internal.VersionInfo

	cmdProfilerBuilder profiling.CmdProfilerBuilder
}

func New(conn net.Conn, backend *backend.Backend, sessionID int, versionInfo *internal.VersionInfo, profiler profiling.CmdProfilerBuilder, eventCh chan<- events.Event) *Session {
	return &Session{
		conn:               conn,
		liner:              liner.New(conn),
		backend:            backend,
		caps:               []imap.Capability{imap.IMAP4rev1, imap.IDLE, imap.UNSELECT, imap.UIDPLUS, imap.MOVE},
		sessionID:          sessionID,
		eventCh:            eventCh,
		version:            versionInfo,
		cmdProfilerBuilder: profiler,
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

	if err := s.greet(); err != nil {
		return err
	}

	if err := s.serve(ctx, s.getCommandCh(ctx, s.backend.GetDelimiter())); err != nil {
		logrus.WithError(err).Errorf("Failed to serve session %v", s.sessionID)
		return err
	}

	return nil
}

func (s *Session) serve(ctx context.Context, cmdCh <-chan command) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	profiler := s.cmdProfilerBuilder.New()
	defer s.cmdProfilerBuilder.Collect(profiler)

	var (
		tag string
		cmd *proto.Command
	)

	for {
		select {
		case res, ok := <-cmdCh:
			if !ok {
				return nil
			}

			tag, cmd = res.tag, res.cmd

		case <-s.state.Done():
			return nil

		case stateUpdate := <-s.state.GetStateUpdatesCh():
			if err := s.state.ApplyUpdate(ctx, stateUpdate); err != nil {
				logrus.WithError(err).Error("Failed to apply state update")
			}

			continue

		case <-ctx.Done():
			return ctx.Err()
		}

		switch {
		case cmd.GetLogout() != nil:
			profiler.Start(profiling.CmdTypeLogout)
			defer profiler.Stop(profiling.CmdTypeLogout)

			return s.handleLogout(ctx, tag, cmd.GetLogout())

		case cmd.GetIdle() != nil:
			profiler.Start(profiling.CmdTypeIdle)
			defer profiler.Stop(profiling.CmdTypeIdle)

			if err := s.handleIdle(ctx, tag, cmd.GetIdle(), cmdCh); err != nil {
				if err := response.No(tag).WithError(err).Send(s); err != nil {
					return fmt.Errorf("failed to send response to client: %w", err)
				}
			}

		default:
			responseCh := s.handleOther(withStartTime(ctx, time.Now()), tag, cmd, profiler)
			for res := range responseCh {
				if err := res.Send(s); err != nil {
					// Consume all remaining channel response since the connection is no longer available.
					// Failing to do so can cause a deadlock in the program as `s.handleOther` never finishes
					// executing and can hold onto a number of locks indefinitely.
					for range responseCh {
						// ...
					}

					return fmt.Errorf("failed to send response to client: %w", err)
				}
			}
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
			logrus.WithError(err).Error("Failed to close state")
		}
	}

	if err := s.conn.Close(); err != nil {
		logrus.WithError(err).Error("Failed to close connection")
	}
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
