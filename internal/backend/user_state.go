package backend

import (
	"context"
	"fmt"

	"github.com/ProtonMail/gluon/internal/backend/ent"
	"github.com/ProtonMail/gluon/internal/remote"
	"github.com/bradenaw/juniper/xslices"
	"golang.org/x/exp/maps"
)

func (user *user) newState(metadataID remote.ConnMetadataID) (*State, error) {
	user.statesLock.Lock()
	defer user.statesLock.Unlock()

	if err := user.remote.CreateConnMetadataStore(metadataID); err != nil {
		return nil, err
	}

	state := &State{user: user, metadataID: metadataID}

	user.states[user.stateID] = state
	user.stateID++

	return state, nil
}

func (user *user) closeState(ctx context.Context, state *State) error {
	return user.tx(ctx, func(tx *ent.Tx) error {
		user.statesLock.Lock()
		defer user.statesLock.Unlock()

		if err := state.deleteConnMetadata(); err != nil {
			return err
		}

		if err := state.close(ctx, tx); err != nil {
			return fmt.Errorf("failed to close state: %w", err)
		}

		delete(user.states, user.stateID)

		return nil
	})
}

func (user *user) closeStates(ctx context.Context) error {
	return user.tx(ctx, func(tx *ent.Tx) error {
		user.statesLock.Lock()
		defer user.statesLock.Unlock()

		for stateID, state := range user.states {
			if err := state.deleteConnMetadata(); err != nil {
				return err
			}

			if err := state.close(ctx, tx); err != nil {
				return fmt.Errorf("failed to close state: %w", err)
			}

			delete(user.states, stateID)
		}

		return nil
	})
}

func (user *user) hasStateInMailboxWithMessage(mboxID, messageID string) bool {
	user.statesLock.RLock()
	defer user.statesLock.RUnlock()

	return xslices.CountFunc(maps.Values(user.states), func(state *State) bool {
		return state.snap != nil && state.snap.mboxID == mboxID && state.snap.hasMessage(messageID)
	}) > 0
}

// forState iterates through all states.
func (user *user) forState(fn func(*State) error) error {
	user.statesLock.RLock()
	defer user.statesLock.RUnlock()

	for _, state := range user.states {
		if err := fn(state); err != nil {
			return err
		}
	}

	return nil
}

// forStateInMailbox iterates through each state that is open in the given mailbox.
func (user *user) forStateInMailbox(mboxID string, fn func(*State) error) error {
	user.statesLock.RLock()
	defer user.statesLock.RUnlock()

	for _, state := range user.states {
		if state.snap != nil && state.snap.mboxID == mboxID {
			if err := fn(state); err != nil {
				return err
			}
		}
	}

	return nil
}

// forStateWithMessage iterates through all states that have the given message.
func (user *user) forStateWithMessage(messageID string, fn func(*State) error) error {
	user.statesLock.RLock()
	defer user.statesLock.RUnlock()

	for _, state := range user.states {
		if state.snap != nil && state.snap.hasMessage(messageID) {
			if err := fn(state); err != nil {
				return err
			}
		}
	}

	return nil
}

// forStateInMailboxWithMessage iterates through each state that is open in the given mailbox which contains the given message.
// A state might still contain the given message if the message had been expunged but this state was not notified yet.
func (user *user) forStateInMailboxWithMessage(mboxID, messageID string, fn func(*State) error) error {
	return user.forStateInMailbox(mboxID, func(state *State) error {
		if state.snap.hasMessage(messageID) {
			return fn(state)
		}

		return nil
	})
}

// use by any other states associated with this user.
// WARNING: This function needs to called from a within a user.stateLock scope.
func (user *user) collectUnusedMessagesMarkedForDelete(ctx context.Context, tx *ent.Tx, excludedState *State) ([]string, error) {
	messageIDsMarkForDelete, err := txGetMessageIDsMarkedDeleted(ctx, tx)
	if err != nil {
		return nil, err
	}

	messageIDsToDelete := xslices.Filter(messageIDsMarkForDelete, func(id string) bool {
		for _, st := range user.states {
			if st == excludedState || st.snap == nil {
				continue
			}
			if st.snap.hasMessage(id) {
				return false
			}
		}

		return true
	})

	return messageIDsToDelete, nil
}

// deleteUnusedMessagesMarkedDeleted Will delete all messages that have been marked for deletion that have are not in
// use by any other states associated with this user.
// WARNING: This function needs to called from a within a user.stateLock scope.
func (user *user) deleteUnusedMessagesMarkedDeleted(ctx context.Context, tx *ent.Tx, state *State) error {
	// Don't run this code if the context has been cancelled (e.g: Server shutdown).
	if ctx.Err() != nil {
		return nil
	}

	messageIDsToDelete, err := user.collectUnusedMessagesMarkedForDelete(ctx, tx, state)
	if err != nil {
		return err
	}

	if err := txDeleteMessages(ctx, tx, messageIDsToDelete...); err != nil {
		return err
	}

	return user.store.Delete(messageIDsToDelete...)
}
