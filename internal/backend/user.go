package backend

import (
	"context"
	"fmt"
	"runtime/pprof"
	"sync"

	"github.com/ProtonMail/gluon/connector"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/contexts"
	"github.com/ProtonMail/gluon/internal/db"
	"github.com/ProtonMail/gluon/internal/db/ent"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/gluon/store"
	"github.com/bradenaw/juniper/xslices"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
)

type user struct {
	userID string

	updateInjector *updateInjector
	connector      connector.Connector
	store          store.Store
	delimiter      string

	db *db.DB

	states      map[state.StateID]*state.State
	statesLock  sync.RWMutex
	nextStateID int

	updateWG     sync.WaitGroup
	updateQuitCh chan struct{}

	// statesWG is
	statesWG sync.WaitGroup
}

func newUser(ctx context.Context, userID string, db *db.DB, conn connector.Connector, store store.Store, delimiter string) (*user, error) {
	if err := db.Init(ctx); err != nil {
		return nil, err
	}

	user := &user{
		userID:         userID,
		connector:      conn,
		updateInjector: newUpdateInjector(ctx, conn, userID),
		store:          store,
		delimiter:      delimiter,
		db:             db,
		states:         make(map[state.StateID]*state.State),
		updateQuitCh:   make(chan struct{}),
	}

	if err := user.deleteAllMessagesMarkedDeleted(ctx); err != nil {
		return nil, err
	}

	user.updateWG.Add(1)

	go func() {
		defer user.updateWG.Done()
		labels := pprof.Labels("go", "Connector Updates", "UserID", user.userID)
		pprof.Do(ctx, labels, func(_ context.Context) {
			ctx := contexts.NewRemoteUpdateCtx(context.Background())
			updateCh := user.updateInjector.GetUpdates()
			for {
				select {
				case update := <-updateCh:
					if err := user.apply(ctx, update); err != nil {
						reporter.MessageWithContext(ctx,
							"Failed to apply connector update",
							reporter.Context{"error": err, "update": update.String()},
						)

						logrus.WithError(err).Errorf("Failed to apply update: %v", err)
					}
				case <-user.updateQuitCh:
					return
				}
			}
		})
	}()

	return user, nil
}

// close closes the backend user.
func (user *user) close(ctx context.Context) error {
	close(user.updateQuitCh)

	// Wait until the connector update go routine has finished.
	user.updateWG.Wait()

	if err := user.updateInjector.Close(ctx); err != nil {
		return err
	}

	if err := user.connector.Close(ctx); err != nil {
		return err
	}

	user.closeStates()

	// Ensure we wait until all states have been removed/closed by any active sessions otherwise we run  into issues
	// since we close the database in this function.
	user.statesWG.Wait()

	if err := user.store.Close(); err != nil {
		return fmt.Errorf("failed to close user client storage: %w", err)
	}

	if err := user.db.Close(); err != nil {
		return fmt.Errorf("failed to close user db: %w", err)
	}

	return nil
}

func (user *user) deleteAllMessagesMarkedDeleted(ctx context.Context) error {
	return db.WriteAndStore(ctx, user.db, user.store, func(ctx context.Context, tx *ent.Tx, stx store.Transaction) error {
		ids, err := db.GetMessageIDsMarkedDeleted(ctx, tx.Client())
		if err != nil {
			return err
		}

		if err := db.DeleteMessages(ctx, tx, ids...); err != nil {
			return err
		}

		return stx.Delete(ids...)
	})
}

func (user *user) queueStateUpdate(update state.Update) {
	if err := user.forState(func(state *state.State) error {
		if !state.QueueUpdates(update) {
			logrus.Errorf("Failed to push update to state %v", state.StateID)
		}
		return nil
	}); err != nil {
		panic("unexpected, should not happen")
	}
}

func (user *user) newState() (*state.State, error) {
	user.statesLock.Lock()
	defer user.statesLock.Unlock()

	user.nextStateID++

	newState := state.NewState(
		state.StateID(user.nextStateID),
		newStateUserInterfaceImpl(user, newStateConnectorImpl(user)),
		user.delimiter,
	)

	user.states[state.StateID(user.nextStateID)] = newState

	user.statesWG.Add(1)

	return newState, nil
}

func (user *user) removeState(ctx context.Context, st *state.State) error {
	messageIDs, err := db.ReadResult(ctx, user.db, func(ctx context.Context, client *ent.Client) ([]imap.InternalMessageID, error) {
		return db.GetMessageIDsMarkedDeleted(ctx, client)
	})
	if err != nil {
		return err
	}

	// We need to reduce the scope of this lock as it can deadlock when there's an IMAP update running
	// at the same time as we remove a state. When the IMAP update propagates the info the order of the locks
	// is inverse to the order we have here.
	fn := func() (*state.State, error) {
		user.statesLock.Lock()
		defer user.statesLock.Unlock()

		st, ok := user.states[st.StateID]
		if !ok {
			return nil, fmt.Errorf("no such state")
		}

		messageIDs = xslices.Filter(messageIDs, func(messageID imap.InternalMessageID) bool {
			return xslices.CountFunc(maps.Values(user.states), func(other *state.State) bool {
				return st != other && other.HasMessage(messageID)
			}) == 0
		})

		delete(user.states, st.StateID)

		return st, nil
	}

	state, err := fn()
	if err != nil {
		return err
	}

	// After this point we need to notify the WaitGroup or we risk deadlocks.
	defer user.statesWG.Done()

	if err := db.WriteAndStore(ctx, user.db, user.store, func(ctx context.Context, tx *ent.Tx, stx store.Transaction) error {
		if err := db.DeleteMessages(ctx, tx, messageIDs...); err != nil {
			return err
		}

		if err := stx.Delete(messageIDs...); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return state.Close(ctx)
}

// forState iterates through all states.
func (user *user) forState(fn func(*state.State) error) error {
	user.statesLock.RLock()
	defer user.statesLock.RUnlock()

	for _, state := range user.states {
		if err := fn(state); err != nil {
			return err
		}
	}

	return nil
}

func (user *user) closeStates() {
	user.statesLock.RLock()
	defer user.statesLock.RUnlock()

	for _, state := range user.states {
		state.SignalClose()
	}
}
