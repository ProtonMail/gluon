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
