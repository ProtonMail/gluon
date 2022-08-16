package backend

import (
	"context"
	"fmt"
	"github.com/ProtonMail/gluon/imap"

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

	user.nextStateID++

	user.states[user.nextStateID] = &State{
		user:       user,
		stateID:    user.nextStateID,
		metadataID: metadataID,
		doneCh:     make(chan struct{}),
		stopCh:     make(chan struct{}),
	}

	return user.states[user.nextStateID], nil
}

func (user *user) removeState(ctx context.Context, stateID int) error {
	return user.tx(ctx, func(tx *ent.Tx) error {
		user.statesLock.Lock()
		defer user.statesLock.Unlock()

		state, ok := user.states[stateID]
		if !ok {
			panic("no such state")
		}

		messageIDs, err := DBGetMessageIDsMarkedDeleted(ctx, tx)
		if err != nil {
			return err
		}

		messageIDs = xslices.Filter(messageIDs, func(messageID imap.InternalMessageID) bool {
			return xslices.CountFunc(maps.Values(user.states), func(other *State) bool {
				return state != other && other.snap != nil && other.snap.hasMessage(messageID)
			}) == 0
		})

		if err := DBDeleteMessages(ctx, tx, messageIDs...); err != nil {
			return err
		}

		if err := user.store.Delete(xslices.Map(messageIDs, func(i imap.InternalMessageID) string {
			return string(i)
		})...); err != nil {
			return err
		}

		if err := state.deleteConnMetadata(); err != nil {
			return fmt.Errorf("failed to remove conn metadata: %w", err)
		}

		if err := state.close(ctx, tx); err != nil {
			return fmt.Errorf("failed to close state: %w", err)
		}

		delete(user.states, stateID)

		return nil
	})
}

func (user *user) hasStateInMailboxWithMessage(mboxID imap.InternalMailboxID, messageID imap.InternalMessageID) bool {
	user.statesLock.RLock()
	defer user.statesLock.RUnlock()

	return xslices.CountFunc(maps.Values(user.states), func(state *State) bool {
		return state.snap != nil && state.snap.mboxID.InternalID == mboxID && state.snap.hasMessage(messageID)
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
func (user *user) forStateInMailbox(mboxID imap.InternalMailboxID, fn func(*State) error) error {
	user.statesLock.RLock()
	defer user.statesLock.RUnlock()

	for _, state := range user.states {
		if state.snap != nil && state.snap.mboxID.InternalID == mboxID {
			if err := fn(state); err != nil {
				return err
			}
		}
	}

	return nil
}

// forStateWithMessage iterates through all states that have the given message.
func (user *user) forStateWithMessage(messageID imap.InternalMessageID, fn func(*State) error) error {
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
func (user *user) forStateInMailboxWithMessage(mboxID imap.InternalMailboxID, messageID imap.InternalMessageID, fn func(*State) error) error {
	return user.forStateInMailbox(mboxID, func(state *State) error {
		if state.snap.hasMessage(messageID) {
			return fn(state)
		}

		return nil
	})
}

func (user *user) getStates() []*State {
	user.statesLock.RLock()
	defer user.statesLock.RUnlock()

	return maps.Values(user.states)
}

func (user *user) closeStates() {
	for _, state := range user.getStates() {
		close(state.doneCh)
		<-state.stopCh
	}
}
