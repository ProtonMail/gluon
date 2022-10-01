// Package gluon implements an IMAP4rev1 (+ extensions) mailserver.
package gluon

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"runtime/pprof"
	"strconv"
	"sync"
	"time"

	"github.com/ProtonMail/gluon/connector"
	"github.com/ProtonMail/gluon/events"
	"github.com/ProtonMail/gluon/internal/backend"
	"github.com/ProtonMail/gluon/internal/session"
	"github.com/ProtonMail/gluon/profiling"
	"github.com/ProtonMail/gluon/queue"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/gluon/store"
	"github.com/ProtonMail/gluon/version"
	"github.com/ProtonMail/gluon/wait"
	"github.com/ProtonMail/gluon/watcher"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

// Server is the gluon IMAP server.
type Server struct {
	// dir holds the path to all of Gluon's data.
	dir string

	// backend provides the server with access to the IMAP backend.
	backend *backend.Backend

	// sessions holds all active IMAP sessions.
	sessions     map[int]*session.Session
	sessionsLock sync.RWMutex

	// serveErrCh collects errors encountered while serving.
	serveErrCh *queue.QueuedChannel[error]

	// serveDoneCh is used to stop the server.
	serveDoneCh chan struct{}

	// serveWG keeps track of serving goroutines.
	serveWG wait.Group

	// nextID holds the ID that will be given to the next session.
	nextID     int
	nextIDLock sync.Mutex

	// inLogger and outLogger are used to log incoming and outgoing IMAP communications.
	inLogger, outLogger io.Writer

	// tlsConfig is used to serve over TLS.
	tlsConfig *tls.Config

	// watchers holds streams of events.
	watchers     []*watcher.Watcher[events.Event]
	watchersLock sync.RWMutex

	// storeBuilder builds message stores.
	storeBuilder store.Builder

	// cmdExecProfBuilder builds command profiling collectors.
	cmdExecProfBuilder profiling.CmdProfilerBuilder

	// versionInfo holds info about the Gluon version.
	versionInfo version.Info

	// reporter is used to report errors to things like Sentry.
	reporter reporter.Reporter

	// idleBulkTime to control how often IDLE responses are sent. 0 means
	// immediate response with no response merging.
	idleBulkTime time.Duration
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

// AddUser creates a new user and generates new unique ID for this user.
// If you have an existing userID, please use LoadUser instead.
func (s *Server) AddUser(ctx context.Context, conn connector.Connector, passphrase []byte) (string, error) {
	userID := s.backend.NewUserID()

	if err := s.LoadUser(ctx, conn, userID, passphrase); err != nil {
		return "", err
	}

	return userID, nil
}

// LoadUser adds an existing user using a previously crated unique user ID.
// If you don't have an existing userID, please use AddUser instead.
func (s *Server) LoadUser(ctx context.Context, conn connector.Connector, userID string, passphrase []byte) error {
	ctx = reporter.NewContextWithReporter(ctx, s.reporter)

	if err := s.backend.AddUser(ctx, userID, conn, passphrase); err != nil {
		return err
	}

	s.publish(events.UserAdded{
		UserID: userID,
	})

	return nil
}

// RemoveUser removes a user from gluon.
func (s *Server) RemoveUser(ctx context.Context, userID string, removeFiles bool) error {
	ctx = reporter.NewContextWithReporter(ctx, s.reporter)

	if err := s.backend.RemoveUser(ctx, userID, removeFiles); err != nil {
		return err
	}

	s.publish(events.UserRemoved{
		UserID: userID,
	})

	return nil
}

// AddWatcher adds a new watcher which watches events of the given types.
// If no types are specified, the watcher watches all events.
func (s *Server) AddWatcher(ofType ...events.Event) <-chan events.Event {
	s.watchersLock.Lock()
	defer s.watchersLock.Unlock()

	watcher := watcher.New(ofType...)

	s.watchers = append(s.watchers, watcher)

	return watcher.GetChannel()
}

// Serve serves connections accepted from the given listener.
// It stops serving when the context is canceled, the listener is closed, or the server is closed.
func (s *Server) Serve(ctx context.Context, l net.Listener) error {
	ctx = reporter.NewContextWithReporter(ctx, s.reporter)

	s.publish(events.ListenerAdded{
		Addr: l.Addr(),
	})

	s.serveWG.Go(func() {
		defer s.publish(events.ListenerRemoved{
			Addr: l.Addr(),
		})

		s.serve(ctx, newConnCh(l))
	})

	return nil
}

// serve handles incoming connections and starts a new goroutine for each.
func (s *Server) serve(ctx context.Context, connCh <-chan net.Conn) {
	var connWG wait.Group
	defer connWG.Wait()

	for {
		select {
		case <-ctx.Done():
			logrus.Debug("Stopping serve, context canceled")
			return

		case <-s.serveDoneCh:
			logrus.Debug("Stopping serve, server stopped")
			return

		case conn, ok := <-connCh:
			if !ok {
				logrus.Debug("Stopping serve, listener closed")
				return
			}

			defer conn.Close()

			connWG.Go(func() {
				session, sessionID := s.addSession(ctx, conn)
				defer s.removeSession(sessionID)

				labels := pprof.Labels("go", "Serve", "SessionID", strconv.Itoa(sessionID))
				pprof.Do(ctx, labels, func(ctx context.Context) {
					if err := session.Serve(ctx); err != nil {
						if !errors.Is(err, net.ErrClosed) {
							s.serveErrCh.Enqueue(err)
						}
					}
				})
			})
		}
	}
}

// GetErrorCh returns the error channel.
func (s *Server) GetErrorCh() <-chan error {
	return s.serveErrCh.GetChannel()
}

// GetVersionInfo returns the version info.
func (s *Server) GetVersionInfo() version.Info {
	return s.versionInfo
}

// GetDataPath returns the path in which gluon stores its data.
func (s *Server) GetDataPath() string {
	return s.dir
}

// Close closes the server.
func (s *Server) Close(ctx context.Context) error {
	ctx = reporter.NewContextWithReporter(ctx, s.reporter)

	// Tell the server to stop serving.
	close(s.serveDoneCh)

	// Wait until all goroutines currently handling connections are done.
	s.serveWG.Wait()

	// Close the backend.
	if err := s.backend.Close(ctx); err != nil {
		return fmt.Errorf("failed to close backend: %w", err)
	}

	// Close the server error channel.
	s.serveErrCh.Close()

	// Close any watchers.
	for _, watcher := range s.watchers {
		watcher.Close()
	}

	return nil
}

func (s *Server) addSession(ctx context.Context, conn net.Conn) (*session.Session, int) {
	s.sessionsLock.Lock()
	defer s.sessionsLock.Unlock()

	nextID := s.getNextID()

	s.sessions[nextID] = session.New(conn, s.backend, nextID, s.versionInfo, s.cmdExecProfBuilder, s.newEventCh(ctx), s.idleBulkTime)

	if s.tlsConfig != nil {
		s.sessions[nextID].SetTLSConfig(s.tlsConfig)
	}

	if s.inLogger != nil {
		s.sessions[nextID].SetIncomingLogger(s.inLogger)
	}

	if s.outLogger != nil {
		s.sessions[nextID].SetOutgoingLogger(s.outLogger)
	}

	s.publish(events.SessionAdded{
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

	s.publish(events.SessionRemoved{
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

	for _, watcher := range s.watchers {
		if watcher.IsWatching(event) {
			if ok := watcher.Send(event); !ok {
				logrus.WithField("event", event).Warn("Failed to send event to watcher")
			}
		}
	}
}

// newConnCh accepts connections from the given listener.
// It returns a channel of all accepted connections which is closed when the listener is closed.
func newConnCh(l net.Listener) <-chan net.Conn {
	connCh := make(chan net.Conn)

	go func() {
		defer close(connCh)

		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}

			connCh <- conn
		}
	}()

	return connCh
}
