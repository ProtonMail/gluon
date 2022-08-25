package backend

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/ProtonMail/gluon/connector"
	"github.com/ProtonMail/gluon/internal/remote"
	"github.com/ProtonMail/gluon/store"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Backend struct {
	// dir is the directory in which backend files should be stored.
	dir string

	// remote manages operations to be performed on the API.
	remote *remote.Manager

	// delim is the server's path delim.
	delim string

	// users holds all registered backend users.
	users     map[string]*user
	usersLock sync.Mutex

	// storeBuilder builds stores for the backend users.
	storeBuilder store.Builder
}

func New(dir string, storeBuilder store.Builder, delim string) (*Backend, error) {
	manager, err := remote.New(filepath.Join(dir, "remote"))
	if err != nil {
		return nil, err
	}

	return &Backend{
		dir:          dir,
		remote:       manager,
		storeBuilder: storeBuilder,
		delim:        delim,
		users:        make(map[string]*user),
	}, nil
}

func (b *Backend) NewUserID() string {
	return uuid.NewString()
}

func (b *Backend) GetDelimiter() string {
	return b.delim
}

func (b *Backend) AddUser(ctx context.Context, userID string, conn connector.Connector, passphrase []byte) error {
	b.usersLock.Lock()
	defer b.usersLock.Unlock()

	store, err := b.storeBuilder.New(filepath.Join(b.dir, "store"), userID, passphrase)
	if err != nil {
		return err
	}

	db, err := b.newDB(userID)
	if err != nil {
		if err := store.Close(); err != nil {
			logrus.WithError(err).Error("Failed to close storage")
		}

		return err
	}

	remote, err := b.remote.AddUser(ctx, userID, conn)
	if err != nil {
		return err
	}

	user, err := newUser(ctx, userID, db, remote, store, b.delim)
	if err != nil {
		return err
	}

	b.users[userID] = user

	return nil
}

func (b *Backend) RemoveUser(ctx context.Context, userID string) error {
	b.usersLock.Lock()
	defer b.usersLock.Unlock()

	user, ok := b.users[userID]
	if !ok {
		return ErrNoSuchUser
	}

	if err := user.close(ctx); err != nil {
		return fmt.Errorf("failed to close backend user: %w", err)
	}

	if err := b.remote.RemoveUser(ctx, userID); err != nil {
		return fmt.Errorf("failed to remove remote user: %w", err)
	}

	delete(b.users, userID)

	return nil
}

func (b *Backend) GetState(username, password string, sessionID int) (*State, error) {
	b.usersLock.Lock()
	defer b.usersLock.Unlock()

	userID, err := b.remote.GetUserID(username, password)
	if err != nil {
		return nil, err
	}

	state, err := b.users[userID].newState(remote.ConnMetadataID(sessionID))
	if err != nil {
		return nil, err
	}

	logrus.
		WithField("userID", userID).
		WithField("username", username).
		Debug("Created new IMAP state")

	return state, nil
}

func (b *Backend) Close(ctx context.Context) error {
	b.usersLock.Lock()
	defer b.usersLock.Unlock()

	for userID, user := range b.users {
		if err := user.close(ctx); err != nil {
			return fmt.Errorf("failed to close backend user (%v): %w", userID, err)
		}

		if err := b.remote.RemoveUser(ctx, userID); err != nil {
			return err
		}

		delete(b.users, userID)
	}

	logrus.Debug("Backend was closed")

	return nil
}
