package backend

import (
	"context"
	"fmt"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/backend/ent"
	"github.com/bradenaw/juniper/xslices"
)

type messageFlagsAddedStateUpdate struct {
	anyMessageIDStateFilter
	flags   imap.FlagSet
	mboxID  MailboxIDPair
	stateID int
}

func newMessageFlagsAddedStateUpdate(flags imap.FlagSet, mboxID MailboxIDPair, messageIDs []imap.InternalMessageID, stateID int) stateUpdate {
	return &messageFlagsAddedStateUpdate{
		flags:                   flags,
		mboxID:                  mboxID,
		anyMessageIDStateFilter: anyMessageIDStateFilter{messageIDs: messageIDs},
		stateID:                 stateID,
	}
}

func (u *messageFlagsAddedStateUpdate) apply(ctx context.Context, tx *ent.Tx, s *State) error {
	for _, messageID := range u.messageIDs {
		snapFlags, err := s.snap.getMessageFlags(messageID)
		if err != nil {
			return err
		}

		newFlags := snapFlags.AddFlagSet(u.flags)

		if s.snap.mboxID != u.mboxID {
			newFlags = newFlags.Set(imap.FlagDeleted, snapFlags.Contains(imap.FlagDeleted))
		}

		if err := s.pushResponder(ctx, tx, newFetch(
			messageID,
			newFlags,
			isUID(ctx),
			s.stateID == u.stateID && isSilent(ctx),
		)); err != nil {
			return err
		}
	}

	return nil
}

// applyMessageFlagsAdded adds the flags to the given messages.
func (state *State) applyMessageFlagsAdded(ctx context.Context, tx *ent.Tx, messageIDs []imap.InternalMessageID, addFlags imap.FlagSet) error {
	if addFlags.Contains(imap.FlagRecent) {
		return fmt.Errorf("the recent flag is read-only")
	}

	client := tx.Client()

	curFlags, err := DBGetMessageFlags(ctx, client, messageIDs)
	if err != nil {
		return err
	}

	delFlags, err := DBGetMessageDeleted(ctx, client, state.snap.mboxID.InternalID, messageIDs)
	if err != nil {
		return err
	}

	if addFlags.Contains(imap.FlagDeleted) {
		if err := DBSetDeletedFlag(ctx, tx, state.snap.mboxID.InternalID, xslices.Filter(messageIDs, func(messageID imap.InternalMessageID) bool {
			return !delFlags[messageID]
		}), true); err != nil {
			return err
		}
	}

	for _, flag := range addFlags.Remove(imap.FlagDeleted).ToSlice() {
		if err := DBAddMessageFlag(ctx, tx, xslices.Filter(messageIDs, func(messageID imap.InternalMessageID) bool {
			return !curFlags[messageID].Contains(flag)
		}), flag); err != nil {
			return err
		}
	}

	if err := state.user.queueOrApplyStateUpdate(ctx, tx, newMessageFlagsAddedStateUpdate(addFlags, state.snap.mboxID, messageIDs, state.stateID)); err != nil {
		return err
	}

	return nil
}

type messageFlagsRemovedStateUpdate struct {
	anyMessageIDStateFilter
	flags   imap.FlagSet
	mboxID  MailboxIDPair
	stateID int
}

func newMessageFlagsRemovedStateUpdate(flags imap.FlagSet, mboxID MailboxIDPair, messageIDs []imap.InternalMessageID, stateID int) stateUpdate {
	return &messageFlagsRemovedStateUpdate{
		flags:                   flags,
		mboxID:                  mboxID,
		anyMessageIDStateFilter: anyMessageIDStateFilter{messageIDs: messageIDs},
		stateID:                 stateID,
	}
}

func (u *messageFlagsRemovedStateUpdate) apply(ctx context.Context, tx *ent.Tx, s *State) error {
	for _, messageID := range u.messageIDs {
		snapFlags, err := s.snap.getMessageFlags(messageID)
		if err != nil {
			return err
		}

		newFlags := snapFlags.RemoveFlagSet(u.flags)

		if s.snap.mboxID != u.mboxID {
			newFlags = newFlags.Set(imap.FlagDeleted, snapFlags.Contains(imap.FlagDeleted))
		}

		if err := s.pushResponder(ctx, tx, newFetch(
			messageID,
			newFlags,
			isUID(ctx),
			s.stateID == u.stateID && isSilent(ctx),
		)); err != nil {
			return err
		}
	}

	return nil
}

// applyMessageFlagsRemoved removes the flags from the given messages.
func (state *State) applyMessageFlagsRemoved(ctx context.Context, tx *ent.Tx, messageIDs []imap.InternalMessageID, remFlags imap.FlagSet) error {
	if remFlags.Contains(imap.FlagRecent) {
		return fmt.Errorf("the recent flag is read-only")
	}

	client := tx.Client()

	curFlags, err := DBGetMessageFlags(ctx, client, messageIDs)
	if err != nil {
		return err
	}

	delFlags, err := DBGetMessageDeleted(ctx, client, state.snap.mboxID.InternalID, messageIDs)
	if err != nil {
		return err
	}

	if remFlags.Contains(imap.FlagDeleted) {
		if err := DBSetDeletedFlag(ctx, tx, state.snap.mboxID.InternalID, xslices.Filter(messageIDs, func(messageID imap.InternalMessageID) bool {
			return delFlags[messageID]
		}), false); err != nil {
			return err
		}
	}

	for _, flag := range remFlags.Remove(imap.FlagDeleted).ToSlice() {
		if err := DBRemoveMessageFlag(ctx, tx, xslices.Filter(messageIDs, func(messageID imap.InternalMessageID) bool {
			return curFlags[messageID].Contains(flag)
		}), flag); err != nil {
			return err
		}
	}

	if err := state.user.queueOrApplyStateUpdate(ctx, tx, newMessageFlagsRemovedStateUpdate(remFlags, state.snap.mboxID, messageIDs, state.stateID)); err != nil {
		return err
	}

	res := make(map[imap.InternalMessageID]imap.FlagSet)

	for _, messageID := range messageIDs {
		res[messageID] = curFlags[messageID].RemoveFlagSet(remFlags)
	}

	return nil
}

type messageFlagsSetStateUpdate struct {
	anyMessageIDStateFilter
	flags   imap.FlagSet
	mboxID  MailboxIDPair
	stateID int
}

func newMessageFlagsSetStateUpdate(flags imap.FlagSet, mboxID MailboxIDPair, messageIDs []imap.InternalMessageID, stateID int) stateUpdate {
	return &messageFlagsSetStateUpdate{
		flags:                   flags,
		mboxID:                  mboxID,
		anyMessageIDStateFilter: anyMessageIDStateFilter{messageIDs: messageIDs},
		stateID:                 stateID,
	}
}

func (u *messageFlagsSetStateUpdate) apply(ctx context.Context, tx *ent.Tx, state *State) error {
	for _, messageID := range u.messageIDs {
		newFlags := u.flags

		if state.snap.mboxID != u.mboxID {
			snapFlags, err := state.snap.getMessageFlags(messageID)
			if err != nil {
				return err
			}

			newFlags = newFlags.Set(imap.FlagDeleted, snapFlags.Contains(imap.FlagDeleted))
		}

		if err := state.pushResponder(ctx, tx, newFetch(
			messageID,
			newFlags,
			isUID(ctx),
			state.stateID == u.stateID && isSilent(ctx),
		)); err != nil {
			return err
		}
	}

	return nil
}

// applyMessageFlagsSet sets the flags of the given messages.
func (state *State) applyMessageFlagsSet(ctx context.Context, tx *ent.Tx, messageIDs []imap.InternalMessageID, setFlags imap.FlagSet) error {
	if setFlags.Contains(imap.FlagRecent) {
		return fmt.Errorf("the recent flag is read-only")
	}

	if state.snap == nil {
		return nil
	}

	if err := DBSetDeletedFlag(ctx, tx, state.snap.mboxID.InternalID, messageIDs, setFlags.Contains(imap.FlagDeleted)); err != nil {
		return err
	}

	if err := DBSetMessageFlags(ctx, tx, messageIDs, setFlags.Remove(imap.FlagDeleted)); err != nil {
		return err
	}

	if err := state.user.queueOrApplyStateUpdate(ctx, tx, newMessageFlagsSetStateUpdate(setFlags, state.snap.mboxID, messageIDs, state.stateID)); err != nil {
		return err
	}

	return nil
}

type mailboxRemoteIDUpdateStateUpdate struct {
	stateFilter
	remoteID imap.LabelID
}

func newMailboxRemoteIDUpdateStateUpdate(internalID imap.InternalMailboxID, remoteID imap.LabelID) stateUpdate {
	return &mailboxRemoteIDUpdateStateUpdate{
		stateFilter: newMBoxIDStateFilter(internalID),
		remoteID:    remoteID,
	}
}

func (u *mailboxRemoteIDUpdateStateUpdate) apply(ctx context.Context, tx *ent.Tx, s *State) error {
	s.snap.mboxID.RemoteID = u.remoteID

	return nil
}
