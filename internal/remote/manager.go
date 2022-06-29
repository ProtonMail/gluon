package remote

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/ProtonMail/gluon/connector"
)

var ErrNoSuchUser = errors.New("no such user")

// Manager provides access to remote users.
type Manager struct {
	// dir is the base directory in which serialized operation queues are stored.
	dir string

	// users holds remote users.
	users     map[string]*User
	usersLock sync.RWMutex
}

// New returns a new manager which stores serialized operation queues for users in the given base directory.
func New(dir string) (*Manager, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}

	return &Manager{
		dir:   dir,
		users: make(map[string]*User),
	}, nil
}

// AddUser adds the remote user with the given (IMAP) credentials to the remote manager.
// The user interacts with the remote via the given connector.
func (m *Manager) AddUser(ctx context.Context, userID string, conn connector.Connector) (*User, error) {
	m.usersLock.Lock()
	defer m.usersLock.Unlock()

	path, err := m.getQueuePath(userID)
	if err != nil {
		return nil, err
	}

	user, err := newUser(ctx, userID, path, conn)
	if err != nil {
		return nil, err
	}

	m.users[userID] = user

	return user, nil
}

// RemoveUser removes the user with the given ID from the remote manager.
// It waits until all the user's queued operations have been performed.
func (m *Manager) RemoveUser(ctx context.Context, userID string) error {
	m.usersLock.Lock()
	defer m.usersLock.Unlock()

	user, ok := m.users[userID]
	if !ok {
		return ErrNoSuchUser
	}

	if err := user.CloseAndFlushOperationQueue(ctx); err != nil {
		return err
	}

	path, err := m.getQueuePath(userID)
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil {
		return err
	}

	delete(m.users, userID)

	return nil
}

// GetUserID returns the user ID of the user with the given credentials.
func (m *Manager) GetUserID(username, password string) (string, error) {
	m.usersLock.Lock()
	defer m.usersLock.Unlock()

	for _, user := range m.users {
		if user.conn.Authorize(username, password) {
			return user.userID, nil
		}
	}

	return "", ErrNoSuchUser
}

// getQueuePath returns a path for the user with the given ID to store its serialized operation queue.
func (m *Manager) getQueuePath(userID string) (string, error) {
	return filepath.Join(m.dir, fmt.Sprintf("%v.queue", userID)), nil
}
