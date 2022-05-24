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
	"github.com/ProtonMail/gluon/internal/backend"
	"github.com/ProtonMail/gluon/internal/liner"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/pkg/errors"
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
	state *backend.State

	// name holds the name of the currently selected mailbox, if any.
	name string

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
}

func New(conn net.Conn, backend *backend.Backend, sessionID int, eventCh chan<- events.Event) *Session {
	return &Session{
		conn:      conn,
		liner:     liner.New(conn),
		backend:   backend,
		caps:      []imap.Capability{imap.IMAP4rev1, imap.IDLE, imap.UNSELECT, imap.UIDPLUS, imap.MOVE},
		sessionID: sessionID,
		eventCh:   eventCh,
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

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := s.greet(); err != nil {
		return err
	}

	for {
		tag, cmd, err := s.readCommand()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			} else if err := response.Bad(tag).WithError(err).Send(s); err != nil {
				return err
			}

			continue
		}

		switch {
		case cmd.GetLogout() != nil:
			return s.handleLogout(ctx, cmd.tag, cmd.GetLogout())

		case cmd.GetStartTLS() != nil:
			if err := s.handleStartTLS(ctx, cmd.tag, cmd.GetStartTLS()); err != nil {
				return response.No(cmd.tag).WithError(err).Send(s)
			}

		case cmd.GetIdle() != nil:
			if err := s.handleIdle(ctx, cmd.tag, cmd.GetIdle()); err != nil {
				if err := response.No(cmd.tag).WithError(err).Send(s); err != nil {
					logrus.WithError(err).Error("Failed to send response to client")
				}
			}

		default:
			for res := range s.handleOther(withStartTime(ctx, time.Now()), cmd) {
				if err := res.Send(s); err != nil {
					logrus.WithError(err).Error("Failed to send response to client")
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

func (s *Session) readCommand() (string, *IMAPCommand, error) {
	line, literals, err := s.liner.Read(func() error { return response.Continuation().Send(s) })
	if err != nil {
		return "", nil, err
	}

	s.logIncoming(string(line))

	tag, cmd, err := parse(line, literals)
	if err != nil {
		logrus.WithError(err).Error("Failed to parse IMAP command")
		return tag, cmd, err
	}

	return tag, cmd, nil
}

func (s *Session) logIncoming(line string) {
	if s.inLogger == nil {
		return
	}

	var name string

	if s.name != "" {
		name = s.name
	} else {
		name = "--"
	}

	writeLog(s.inLogger, "C", strconv.Itoa(s.sessionID), name, line)
}

func (s *Session) logOutgoing(line string) {
	if s.outLogger == nil {
		return
	}

	var name string

	if s.name != "" {
		name = s.name
	} else {
		name = "--"
	}

	writeLog(s.outLogger, "S", strconv.Itoa(s.sessionID), name, line)
}

func (s *Session) done(ctx context.Context) {
	if s.state != nil {
		if err := s.state.Close(ctx); err != nil {
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
		WithMessage(fmt.Sprintf("gluon session ID %v", s.sessionID)).
		Send(s)
}
