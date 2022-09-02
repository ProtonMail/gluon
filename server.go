// Package gluon implements an IMAP4rev1 (+ extensions) mailserver.
package gluon

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"path/filepath"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync"

	"github.com/ProtonMail/gluon/connector"
	"github.com/ProtonMail/gluon/events"
	"github.com/ProtonMail/gluon/internal"
	"github.com/ProtonMail/gluon/internal/backend"
	"github.com/ProtonMail/gluon/internal/session"
	"github.com/ProtonMail/gluon/profiling"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/gluon/store"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

// Server is the gluon IMAP server.
type Server struct {
	// dir holds the path to all of Gluon's data.
	dir string

	// backend provides the server with access to the IMAP backend.
	backend *backend.Backend

	// listeners holds all listeners on which the server is listening.
	listeners     map[net.Listener]struct{}
	listenersLock sync.Mutex

	// sessions holds all active IMAP sessions.
	sessions     map[int]*session.Session
	sessionsLock sync.RWMutex

	// nextID holds the ID that will be given to the next session.
	nextID     int
	nextIDLock sync.Mutex

	// inLogger and outLogger are used to log incoming and outgoing IMAP communications.
	inLogger, outLogger io.Writer

	// tlsConfig is used to serve over TLS.
	tlsConfig *tls.Config

	// watchers holds streams of events.
	watchers     map[chan events.Event]struct{}
	watchersLock sync.RWMutex

	// storeBuilder builds message stores.
	storeBuilder store.Builder

	// cmdExecProfBuilder builds command profiling collectors.
	cmdExecProfBuilder profiling.CmdProfilerBuilder

	// versionInfo holds info about the Gluon version.
	versionInfo internal.VersionInfo

	connectionWG sync.WaitGroup

	reporter reporter.Reporter
}

// New creates a new server with the given options.
func New(withOpt ...Option) (*Server, error) {
	builder, err := newBuilder()
	if err != nil {
		return nil, err
	}

	for _, opt := range withOpt {
		opt.config(builder)
	}

	return builder.build()
}

// AddUser creates a new user and generates new unique ID for this user. If you have an existing userID, please use
// LoadUser instead.
func (s *Server) AddUser(ctx context.Context, conn connector.Connector, encryptionPassphrase []byte) (string, error) {
	userID := s.backend.NewUserID()

	if err := s.LoadUser(ctx, conn, userID, encryptionPassphrase); err != nil {
		return "", err
	}

	return userID, nil
}

// LoadUser loads an existing user's data from disk. This function can also be used to assign a custom userID to a mail
// server user.
func (s *Server) LoadUser(ctx context.Context, conn connector.Connector, userID string, passphrase []byte) error {
	ctx = reporter.NewContextWithReporter(ctx, s.reporter)

	if err := s.backend.AddUser(ctx, userID, conn, passphrase); err != nil {
		return err
	}

	s.publish(events.EventUserAdded{
		UserID: userID,
	})

	return nil
}

// RemoveUser removes a user from the mailserver.
func (s *Server) RemoveUser(ctx context.Context, userID string) error {
	ctx = reporter.NewContextWithReporter(ctx, s.reporter)

	if err := s.backend.RemoveUser(ctx, userID); err != nil {
		return err
	}

	s.publish(events.EventUserRemoved{
		UserID: userID,
	})

	return nil
}

// AddWatcher adds a new watcher.
func (s *Server) AddWatcher() chan events.Event {
	s.watchersLock.Lock()
	defer s.watchersLock.Unlock()

	eventCh := make(chan events.Event)

	s.watchers[eventCh] = struct{}{}

	return eventCh
}

// RemoveWatcher removes the watcher from the server and closes the channel.
func (s *Server) RemoveWatcher(ch chan events.Event) {
	s.watchersLock.Lock()
	defer s.watchersLock.Unlock()

	if _, ok := s.watchers[ch]; ok {
		close(ch)
		delete(s.watchers, ch)
	}
}

// Serve serves connections accepted from the given listener.
// It returns a channel of all errors which occur while serving.
// The error channel is closed when either the connection is dropped or the server is closed.
func (s *Server) Serve(ctx context.Context, l net.Listener) chan error {
	errCh := make(chan error)

	ctx = reporter.NewContextWithReporter(ctx, s.reporter)

	go func() {
		defer close(errCh)

		s.addListener(l)
		defer s.removeListener(l)

		defer s.connectionWG.Wait()

		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}

			s.connectionWG.Add(1)

			go func() {
				defer s.connectionWG.Done()
				s.handleConn(ctx, conn, errCh)
			}()
		}
	}()

	return errCh
}

// Close closes the server.
// It firstly closes all TCP listeners then closes the backend.
func (s *Server) Close(ctx context.Context) error {
	ctx = reporter.NewContextWithReporter(ctx, s.reporter)

	for l := range s.listeners {
		s.removeListener(l)
	}

	s.connectionWG.Wait()

	if err := s.backend.Close(ctx); err != nil {
		return fmt.Errorf("failed to close backend: %w", err)
	}

	logrus.Debug("Mailserver was closed")

	return nil
}

func (s *Server) GetVersionInfo() internal.VersionInfo {
	return s.versionInfo
}

func (s *Server) GetDataPath() string {
	return s.dir
}

func (s *Server) GetUserDataPath(userID string) (string, error) {
	if strings.ContainsAny(userID, "./\\") {
		return "", fmt.Errorf("not a valid user id")
	}

	return filepath.Join(s.dir, userID), nil
}

func (s *Server) addListener(l net.Listener) {
	s.listenersLock.Lock()
	defer s.listenersLock.Unlock()

	s.listeners[l] = struct{}{}

	s.publish(events.EventListenerAdded{
		Addr: l.Addr(),
	})
}

func (s *Server) removeListener(l net.Listener) {
	s.listenersLock.Lock()
	defer s.listenersLock.Unlock()

	if _, ok := s.listeners[l]; ok {
		delete(s.listeners, l)

		if err := l.Close(); err != nil {
			logrus.WithError(err).Error("Failed to close listener")
		}
	}

	s.publish(events.EventListenerRemoved{
		Addr: l.Addr(),
	})
}

func (s *Server) handleConn(ctx context.Context, conn net.Conn, errCh chan error) {
	session, sessionID := s.addSession(ctx, conn)
	labels := pprof.Labels("go", "Serve", "SessionID", strconv.Itoa(sessionID))
	pprof.Do(ctx, labels, func(_ context.Context) {
		defer s.removeSession(sessionID)

		if err := session.Serve(ctx); err != nil {
			errCh <- err
		}
	})
}

func (s *Server) addSession(ctx context.Context, conn net.Conn) (*session.Session, int) {
	s.sessionsLock.Lock()
	defer s.sessionsLock.Unlock()

	nextID := s.getNextID()

	s.sessions[nextID] = session.New(conn, s.backend, nextID, &s.versionInfo, s.cmdExecProfBuilder, s.newEventCh(ctx))

	if s.tlsConfig != nil {
		s.sessions[nextID].SetTLSConfig(s.tlsConfig)
	}

	if s.inLogger != nil {
		s.sessions[nextID].SetIncomingLogger(s.inLogger)
	}

	if s.outLogger != nil {
		s.sessions[nextID].SetOutgoingLogger(s.outLogger)
	}

	s.publish(events.EventSessionAdded{
		SessionID:  nextID,
		LocalAddr:  conn.LocalAddr(),
		RemoteAddr: conn.RemoteAddr(),
	})

	return s.sessions[nextID], nextID
}

func (s *Server) removeSession(sessionID int) {
	s.sessionsLock.Lock()
	defer s.sessionsLock.Unlock()

	delete(s.sessions, sessionID)

	s.publish(events.EventSessionRemoved{
		SessionID: sessionID,
	})
}

func (s *Server) getNextID() int {
	s.nextIDLock.Lock()
	defer s.nextIDLock.Unlock()

	s.nextID++

	return s.nextID
}

func (s *Server) newEventCh(ctx context.Context) chan events.Event {
	eventCh := make(chan events.Event)

	go func() {
		labels := pprof.Labels("Server", "Event Channel")
		pprof.Do(ctx, labels, func(_ context.Context) {
			for event := range eventCh {
				s.publish(event)
			}
		})
	}()

	return eventCh
}

func (s *Server) publish(event events.Event) {
	s.watchersLock.RLock()
	defer s.watchersLock.RUnlock()

	for eventCh := range s.watchers {
		eventCh <- event
	}
}
