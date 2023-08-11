package backend

import (
	"context"
	"fmt"
	"sync"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/gluon/connector"
	"github.com/ProtonMail/gluon/db"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/ids"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/ProtonMail/gluon/internal/utils"
	"github.com/ProtonMail/gluon/limits"
	"github.com/ProtonMail/gluon/logging"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/gluon/store"
	"github.com/bradenaw/juniper/xslices"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
)

type user struct {
	userID string

	connector      connector.Connector
	updateInjector *updateInjector
	store          *store.WriteControlledStore
	delimiter      string

	db db.Client

	states     map[state.StateID]*state.State
	statesLock sync.RWMutex
	statesWG   sync.WaitGroup

	updateWG     sync.WaitGroup
	updateQuitCh chan struct{}

	recoveryMailboxID imap.InternalMailboxID

	imapLimits limits.IMAP

	uidValidityGenerator imap.UIDValidityGenerator

	panicHandler async.PanicHandler

	recoveredMessageHashes *utils.MessageHashesMap
}

func newUser(
	ctx context.Context,
	userID string,
	database db.Client,
	conn connector.Connector,
	st store.Store,
	delimiter string,
	imapLimits limits.IMAP,
	uidValidityGenerator imap.UIDValidityGenerator,
	panicHandler async.PanicHandler,
) (*user, error) {
	recoveredMessageHashes := utils.NewMessageHashesMap()

	cacheProvider := NewDBIMAPState(database)

	// Create recovery mailbox if it does not exist
	recoveryMBox, err := db.ClientWriteType(ctx, database, func(ctx context.Context, tx db.Transaction) (*db.Mailbox, error) {
		uidValidity, err := uidValidityGenerator.Generate()
		if err != nil {
			return nil, err
		}

		mboxFlags := imap.NewFlagSet(imap.FlagSeen, imap.FlagFlagged, imap.FlagDeleted)
		mbox := imap.Mailbox{
			ID:             ids.GluonInternalRecoveryMailboxRemoteID,
			Name:           []string{ids.GluonRecoveryMailboxName},
			Flags:          mboxFlags,
			PermanentFlags: mboxFlags,
			Attributes:     imap.NewFlagSet(imap.AttrNoInferiors),
		}

		recoveryMBox, err := tx.GetOrCreateMailboxAlt(ctx, mbox, delimiter, uidValidity)
		if err != nil {
			return nil, err
		}

		// Pre-fill the message hashes map
		messages, err := tx.GetMailboxMessageIDPairs(ctx, recoveryMBox.ID)
		if err != nil {
			return nil, err
		}

		for _, m := range messages {
			literal, err := st.Get(m.InternalID)
			if err != nil {
				logrus.WithError(err).Errorf("Failed to load %v for store for recovered message hashes map", m.InternalID)
				continue
			}

			if _, err := recoveredMessageHashes.Insert(m.InternalID, literal); err != nil {
				logrus.WithError(err).Errorf("Failed insert literal for %v into recovered message hashes map", m.InternalID)
			}
		}

		return recoveryMBox, nil
	})
	if err != nil {
		return nil, err
	}

	if err := conn.Init(ctx, cacheProvider); err != nil {
		logrus.WithError(err).Errorf("Failed to init connector")
		return nil, err
	}

	user := &user{
		userID: userID,

		connector:      conn,
		updateInjector: newUpdateInjector(conn, userID, panicHandler),
		store:          store.NewWriteControlledStore(st),
		delimiter:      delimiter,

		db: database,

		states:       make(map[state.StateID]*state.State),
		updateQuitCh: make(chan struct{}),

		recoveryMailboxID: recoveryMBox.ID,

		imapLimits: imapLimits,

		uidValidityGenerator: uidValidityGenerator,

		panicHandler: panicHandler,

		recoveredMessageHashes: recoveredMessageHashes,
	}

	if err := user.deleteAllMessagesMarkedDeleted(ctx); err != nil {
		logrus.WithError(err).Error("Failed to remove deleted messages")
		reporter.MessageWithContext(ctx,
			"Failed to remove deleted messages",
			reporter.Context{"error": err},
		)
	}

	if err := user.cleanupStaleStoreData(ctx); err != nil {
		logrus.WithError(err).Error("Failed to cleanup stale store data")
	}

	user.updateWG.Add(1)

	// nolint:contextcheck
	async.GoAnnotated(context.Background(), panicHandler, func(ctx context.Context) {
		defer user.updateWG.Done()

		updateCh := user.updateInjector.GetUpdates()

		for {
			select {
			case update, ok := <-updateCh:
				if !ok {
					return
				}

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
	}, logging.Labels{
		"Action": "Applying connector updates",
		"UserID": userID,
	})

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
	// Delete messages in database first before deleting from the storage to avoid data loss.
	ids, err := db.ClientWriteType(ctx, user.db, func(ctx context.Context, tx db.Transaction) ([]imap.InternalMessageID, error) {
		ids, err := tx.GetMessageIDsMarkedAsDelete(ctx)
		if err != nil {
			return nil, err
		}

		if err := tx.DeleteMessages(ctx, ids); err != nil {
			return nil, err
		}

		return ids, nil
	})
	if err != nil {
		return err
	}

	return user.store.Delete(ids...)
}

func (user *user) queueStateUpdate(updates ...state.Update) {
	if err := user.forState(func(state *state.State) error {
		for _, update := range updates {
			if !state.QueueUpdates(update) {
				logrus.Errorf("Failed to push update to state %v", state.StateID)
			}
		}

		return nil
	}); err != nil {
		panic("unexpected, should not happen")
	}
}

func (user *user) newState() (*state.State, error) {
	user.statesLock.Lock()
	defer user.statesLock.Unlock()

	newState := state.NewState(
		newStateUserInterfaceImpl(user, newStateConnectorImpl(user)),
		user.delimiter,
		user.imapLimits,
		user.panicHandler,
	)

	user.states[newState.StateID] = newState

	user.statesWG.Add(1)

	return newState, nil
}

func (user *user) removeState(ctx context.Context, st *state.State) error {
	messageIDs, err := db.ClientReadType(ctx, user.db, func(ctx context.Context, client db.ReadOnly) ([]imap.InternalMessageID, error) {
		return client.GetMessageIDsMarkedAsDelete(ctx)
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

	// Delete messages in database first before deleting from the storage to avoid data loss.
	if err := user.db.Write(ctx, func(ctx context.Context, tx db.Transaction) error {
		if err := tx.DeleteMessages(ctx, messageIDs); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	// If we fail to delete messages on disk, it shouldn't count as an error at this point.
	if err := user.store.Delete(messageIDs...); err != nil {
		logrus.WithError(err).Error("Failed to delete messages during removeState")
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

func (user *user) cleanupStaleStoreData(ctx context.Context) error {
	storeIds, err := user.store.List()
	if err != nil {
		return err
	}

	dbIdMap, err := db.ClientReadType(ctx, user.db, func(ctx context.Context, client db.ReadOnly) (map[imap.InternalMessageID]struct{}, error) {
		return client.GetAllMessagesIDsAsMap(ctx)
	})
	if err != nil {
		return err
	}

	idsToDelete := xslices.Filter(storeIds, func(id imap.InternalMessageID) bool {
		_, ok := dbIdMap[id]

		return !ok
	})

	return user.store.Delete(idsToDelete...)
}
