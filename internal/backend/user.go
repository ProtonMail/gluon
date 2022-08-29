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

	updateWG     sync.WaitGroup
	updateQuitCh chan struct{}
}

func newUser(ctx context.Context, userID string, db *DB, remote *remote.User, store store.Store, delimiter string) (*user, error) {
	if err := db.Init(ctx); err != nil {
		return nil, err
	}

	user := &user{
		userID:       userID,
		remote:       remote,
		store:        store,
		delimiter:    delimiter,
		db:           db,
		states:       make(map[int]*State),
		updateQuitCh: make(chan struct{}),
	}

	if err := user.deleteAllMessagesMarkedDeleted(ctx); err != nil {
		return nil, err
	}

	user.updateWG.Add(1)

	go func() {
		defer user.updateWG.Done()
		labels := pprof.Labels("go", "Connector Updates", "UserID", user.userID)
		pprof.Do(ctx, labels, func(_ context.Context) {
			ctx := NewRemoteUpdateCtx(context.Background())
			updateCh := remote.GetUpdates()
			for {
				select {
				case update := <-updateCh:
					if err := user.apply(ctx, update); err != nil {
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
	user.closeStates()

	close(user.updateQuitCh)

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
	return user.db.Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		ids, err := DBGetMessageIDsMarkedDeleted(ctx, tx.Client())
		if err != nil {
			return err
		}

		if err := DBDeleteMessages(ctx, tx, ids...); err != nil {
			return err
		}

		return user.store.Delete(ids...)
	})
}

func (user *user) queueOrApplyStateUpdate(ctx context.Context, tx *ent.Tx, update stateUpdate) error {
	// If we detect a state id in the context, it means this function call is a result of a User interaction.
	// When that happens the update needs to be applied to the state matching the state ID immediately. If no such
	// stateID exists or the context information is not present, all updates are queued for later execution.
	stateID, ok := getStateIDFromContext(ctx)
	if !ok {
		return user.forState(func(state *State) error {
			state.queueUpdates(update)
			return nil
		})
	} else {
		return user.forState(func(state *State) error {
			if state.stateID != stateID {
				state.queueUpdates(update)

				return nil
			} else {
				if !update.filter(state) {
					return nil
				}

				return update.apply(ctx, tx, state)
			}
		})
	}
}

// stateUserAccessor should be used to interface with the user type from a State type. This is meant to control
// the API boundary layer.
type stateUserAccessor struct {
	u *user
}

func newStateUserAccessor(u *user) stateUserAccessor {
	return stateUserAccessor{u: u}
}

func (s *stateUserAccessor) getUserID() string {
	return s.u.userID
}

func (s *stateUserAccessor) getDelimiter() string {
	return s.u.delimiter
}

func (s *stateUserAccessor) getDB() *DB {
	return s.u.db
}

func (s *stateUserAccessor) removeState(ctx context.Context, stateID int) error {
	return s.u.removeState(ctx, stateID)
}

func (s *stateUserAccessor) getRemote() *remote.User {
	return s.u.remote
}

func (s *stateUserAccessor) getStore() store.Store {
	return s.u.store
}

func (s *stateUserAccessor) applyMessagesAddedToMailbox(
	ctx context.Context,
	tx *ent.Tx,
	mboxID imap.InternalMailboxID,
	messageIDs []imap.InternalMessageID,
) (map[imap.InternalMessageID]int, error) {
	return s.u.applyMessagesAddedToMailbox(ctx, tx, mboxID, messageIDs)
}

func (s *stateUserAccessor) applyMessagesRemovedFromMailbox(ctx context.Context,
	tx *ent.Tx,
	mboxID imap.InternalMailboxID,
	messageIDs []imap.InternalMessageID,
) error {
	return s.u.applyMessagesRemovedFromMailbox(ctx, tx, mboxID, messageIDs)
}

func (s *stateUserAccessor) applyMessagesMovedFromMailbox(
	ctx context.Context,
	tx *ent.Tx,
	mboxFromID, mboxToID imap.InternalMailboxID,
	messageIDs []imap.InternalMessageID,
) (map[imap.InternalMessageID]int, error) {
	return s.u.applyMessagesMovedFromMailbox(ctx, tx, mboxFromID, mboxToID, messageIDs)
}

func (s *stateUserAccessor) queueOrApplyStateUpdate(ctx context.Context, tx *ent.Tx, update stateUpdate) error {
	return s.u.queueOrApplyStateUpdate(ctx, tx, update)
}
