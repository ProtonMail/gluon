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

	newState := NewState(user.nextStateID, metadataID, newStateUserAccessor(user), user.delimiter)

	user.states[user.nextStateID] = newState

	return newState, nil
}

func (user *user) removeState(ctx context.Context, stateID int) error {
	messageIDs, err := DBReadResult(ctx, user.db, func(ctx context.Context, client *ent.Client) ([]imap.InternalMessageID, error) {
		return DBGetMessageIDsMarkedDeleted(ctx, client)
	})
	if err != nil {
		return err
	}

	// We need to reduce the scope of this lock as it can deadlock when there's an IMAP update running
	// at the same time as we remove a state. When the IMAP update propagates the info the order of the locks
	// is inverse to the order we have here.
	fn := func() (*State, error) {
		user.statesLock.Lock()
		defer user.statesLock.Unlock()

		state, ok := user.states[stateID]
		if !ok {
			return nil, fmt.Errorf("no such state")
		}

		messageIDs = xslices.Filter(messageIDs, func(messageID imap.InternalMessageID) bool {
			return xslices.CountFunc(maps.Values(user.states), func(other *State) bool {
				return state != other && other.snap != nil && other.snap.hasMessage(messageID)
			}) == 0
		})

		delete(user.states, stateID)

		return state, nil
	}

	state, err := fn()
	if err != nil {
		return err
	}

	if err := user.db.Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		return DBDeleteMessages(ctx, tx, messageIDs...)
	}); err != nil {
		return err
	}

	if err := user.store.Delete(messageIDs...); err != nil {
		return err
	}

	if err := state.deleteConnMetadata(); err != nil {
		return fmt.Errorf("failed to remove conn metadata: %w", err)
	}

	state.closeUpdateQueue()

	if err := state.close(); err != nil {
		return fmt.Errorf("failed to close state: %w", err)
	}

	return nil
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
