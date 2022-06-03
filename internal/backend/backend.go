package backend

import (
	"context"
	"fmt"
	"sync"

	"github.com/ProtonMail/gluon/connector"
	"github.com/ProtonMail/gluon/internal/backend/ent"
	"github.com/ProtonMail/gluon/internal/remote"
	"github.com/ProtonMail/gluon/store"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Backend struct {
	// remote manages operations to be performed on the API.
	remote *remote.Manager

	// delim is the server's path delim.
	delim string

	// users holds all registered backend users.
	users     map[string]*user
	usersLock sync.Mutex
}

func New(dir string) *Backend {
	return &Backend{
		remote: remote.New(dir),
		delim:  "/",
		users:  make(map[string]*user),
	}
}

func (b *Backend) SetDelimiter(delim string) {
	b.delim = delim
}

func (b *Backend) AddUser(conn connector.Connector, store store.Store, client *ent.Client) (string, error) {
	b.usersLock.Lock()
	defer b.usersLock.Unlock()

	userID := uuid.NewString()

	remote, err := b.remote.AddUser(userID, conn)
	if err != nil {
		return "", err
	}

	user, err := newUser(userID, client, remote, store, b.delim)
	if err != nil {
		return "", err
	}

	b.users[userID] = user

	return userID, nil
}

func (b *Backend) GetState(username, password string) (*State, error) {
	b.usersLock.Lock()
	defer b.usersLock.Unlock()

	userID, err := b.remote.GetUserID(username, password)
	if err != nil {
		return nil, err
	}

	logrus.
		WithField("userID", userID).
		WithField("username", username).
		Debug("Creating new IMAP state")

	return b.users[userID].newState()
}

func (b *Backend) Close(ctx context.Context) error {
	b.usersLock.Lock()
	defer b.usersLock.Unlock()

	for _, user := range b.users {
		if err := user.close(ctx); err != nil {
			return fmt.Errorf("failed to close backend user: %w", err)
		}
	}

	logrus.Debug("Backend was closed")

	return nil
}
