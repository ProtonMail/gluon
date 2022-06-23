package backend

import (
	"context"
	"fmt"
	"sync"

	"github.com/ProtonMail/gluon/internal/backend/ent"
	"github.com/ProtonMail/gluon/internal/remote"
	"github.com/ProtonMail/gluon/store"
	"github.com/sirupsen/logrus"
)

type user struct {
	userID string

	remote    *remote.User
	store     store.Store
	delimiter string

	client *ent.Client
	txLock sync.Mutex

	states     map[int]*State
	statesLock sync.RWMutex
	stateID    int
}

func newUser(userID string, client *ent.Client, remote *remote.User, store store.Store, delimiter string) (*user, error) {
	if err := client.Schema.Create(context.Background()); err != nil {
		return nil, err
	}

	user := &user{
		userID:    userID,
		remote:    remote,
		store:     store,
		delimiter: delimiter,
		client:    client,
		states:    make(map[int]*State),
	}

	go func() {
		for update := range remote.GetUpdates() {
			update := update

			if err := user.tx(context.Background(), func(tx *ent.Tx) error {
				defer update.Done()
				return user.apply(context.Background(), tx, update)
			}); err != nil {
				logrus.WithError(err).Error("Failed to apply update")
			}
		}
	}()

	return user, nil
}

// tx is a helper function that runs a sequence of ent client calls in a transaction.
func (user *user) tx(ctx context.Context, fn func(tx *ent.Tx) error) error {
	user.txLock.Lock()
	defer user.txLock.Unlock()

	tx, err := user.client.Tx(ctx)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			panic(rerr)
		}

		return err
	}

	return tx.Commit()
}

// close closes the backend user.
func (user *user) close(ctx context.Context) error {
	if err := user.closeStates(ctx); err != nil {
		return fmt.Errorf("failed to close user states: %w", err)
	}

	if err := user.remote.CloseAndSerializeOperationQueue(); err != nil {
		return fmt.Errorf("failed to close user remote: %w", err)
	}

	if err := user.client.Close(); err != nil {
		return fmt.Errorf("failed to close user client: %w", err)
	}

	return nil
}
