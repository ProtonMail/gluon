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

	states      map[int]*State
	statesLock  sync.RWMutex
	nextStateID int
}

func newUser(ctx context.Context, userID string, client *ent.Client, remote *remote.User, store store.Store, delimiter string) (*user, error) {
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

	if err := user.deleteAllMessagesMarkedDeleted(ctx); err != nil {
		return nil, err
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

	defer func() {
		if v := recover(); v != nil {
			if err := tx.Rollback(); err != nil {
				panic(fmt.Errorf("rolling back while recovering (%v): %w", v, err))
			}

			panic(v)
		}
	}()

	if err := fn(tx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			return fmt.Errorf("rolling back transaction: %w", rerr)
		}

		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

// close closes the backend user.
func (user *user) close(ctx context.Context) error {
	user.closeStates()

	if err := user.remote.CloseAndSerializeOperationQueue(); err != nil {
		return fmt.Errorf("failed to close user remote: %w", err)
	}

	if err := user.client.Close(); err != nil {
		return fmt.Errorf("failed to close user client: %w", err)
	}

	return nil
}

func (user *user) deleteAllMessagesMarkedDeleted(ctx context.Context) error {
	return user.tx(ctx, func(tx *ent.Tx) error {
		ids, err := txGetMessageIDsMarkedDeleted(ctx, tx)
		if err != nil {
			return err
		}

		if err := txDeleteMessages(ctx, tx, ids...); err != nil {
			return err
		}

		return user.store.Delete(ids...)
	})
}
