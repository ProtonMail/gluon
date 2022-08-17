package backend

import (
	"context"
	"fmt"
	"runtime/pprof"
	"sync"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/backend/ent"
	"github.com/ProtonMail/gluon/internal/remote"
	"github.com/ProtonMail/gluon/store"
	"github.com/bradenaw/juniper/xslices"
	"github.com/sirupsen/logrus"
)

type user struct {
	userID string

	remote    *remote.User
	store     store.Store
	delimiter string

	db *DB

	states      map[int]*State
	statesLock  sync.RWMutex
	nextStateID int

	updateWG sync.WaitGroup
}

func newUser(ctx context.Context, userID string, db *DB, remote *remote.User, store store.Store, delimiter string) (*user, error) {
	if err := db.Init(ctx); err != nil {
		return nil, err
	}

	user := &user{
		userID:    userID,
		remote:    remote,
		store:     store,
		delimiter: delimiter,
		db:        db,
		states:    make(map[int]*State),
	}

	if err := user.deleteAllMessagesMarkedDeleted(ctx); err != nil {
		return nil, err
	}

	user.updateWG.Add(1)

	go func() {
		defer user.updateWG.Done()
		labels := pprof.Labels("go", "Connector Updates", "UserID", user.userID)
		pprof.Do(ctx, labels, func(_ context.Context) {
			for update := range remote.GetUpdates() {
				update := update
				if err := user.tx(context.Background(), func(tx *ent.Tx) error {
					defer update.Done()
					return user.apply(context.Background(), tx, update)
				}); err != nil {
					logrus.WithError(err).Errorf("Failed to apply update: %v", update)
				}
			}
		})
	}()

	return user, nil
}

// tx is a helper function that runs a sequence of ent client calls in a transaction.
func (user *user) tx(ctx context.Context, fn func(tx *ent.Tx) error) error {
	return user.db.Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		return fn(tx)
	})
}

// close closes the backend user.
func (user *user) close(ctx context.Context) error {
	user.closeStates()

	if err := user.remote.Close(ctx); err != nil {
		return fmt.Errorf("failed to close user remote: %w", err)
	}

	// Wait until the connector update go routine has finished.
	user.updateWG.Wait()

	if err := user.store.Close(); err != nil {
		return fmt.Errorf("failed to close user client storage: %w", err)
	}

	if err := user.db.Close(); err != nil {
		return fmt.Errorf("failed to close user db: %w", err)
	}

	return nil
}

func (user *user) deleteAllMessagesMarkedDeleted(ctx context.Context) error {
	return user.tx(ctx, func(tx *ent.Tx) error {
		ids, err := DBGetMessageIDsMarkedDeleted(ctx, tx)
		if err != nil {
			return err
		}

		if err := DBDeleteMessages(ctx, tx, ids...); err != nil {
			return err
		}

		return user.store.Delete(xslices.Map(ids, func(i imap.InternalMessageID) string {
			return string(i)
		})...)
	})
}
