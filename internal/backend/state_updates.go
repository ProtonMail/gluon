package backend

import (
	"context"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/backend/ent"
	"github.com/bradenaw/juniper/xslices"
)

// applyMessageFlagsAdded adds the flags to the given messages.
func (state *State) applyMessageFlagsAdded(ctx context.Context, tx *ent.Tx, messageIDs []imap.InternalMessageID, addFlags imap.FlagSet) error {
	if addFlags.Contains(imap.FlagRecent) {
		panic("the recent flag is read-only")
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

	for _, messageID := range messageIDs {
		if err := state.forStateWithMessage(messageID, func(other *State) error {
			snapFlags, err := other.snap.getMessageFlags(messageID)
			if err != nil {
				return err
			}

			newFlags := snapFlags.AddFlagSet(addFlags)

			if other.snap.mboxID != state.snap.mboxID {
				newFlags = newFlags.Set(imap.FlagDeleted, snapFlags.Contains(imap.FlagDeleted))
			}

			return other.pushResponder(ctx, tx, newFetch(
				messageID,
				newFlags,
				isUID(ctx),
				other == state && isSilent(ctx),
			))
		}); err != nil {
			return err
		}
	}

	return nil
}

// applyMessageFlagsRemoved removes the flags from the given messages.
func (state *State) applyMessageFlagsRemoved(ctx context.Context, tx *ent.Tx, messageIDs []imap.InternalMessageID, remFlags imap.FlagSet) error {
	if remFlags.Contains(imap.FlagRecent) {
		panic("the recent flag is read-only")
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

	for _, messageID := range messageIDs {
		if err := state.forStateWithMessage(messageID, func(other *State) error {
			snapFlags, err := other.snap.getMessageFlags(messageID)
			if err != nil {
				return err
			}

			newFlags := snapFlags.RemoveFlagSet(remFlags)

			if other.snap.mboxID != state.snap.mboxID {
				newFlags = newFlags.Set(imap.FlagDeleted, snapFlags.Contains(imap.FlagDeleted))
			}

			return other.pushResponder(ctx, tx, newFetch(
				messageID,
				newFlags,
				isUID(ctx),
				other == state && isSilent(ctx),
			))
		}); err != nil {
			return err
		}
	}

	res := make(map[imap.InternalMessageID]imap.FlagSet)

	for _, messageID := range messageIDs {
		res[messageID] = curFlags[messageID].RemoveFlagSet(remFlags)
	}

	return nil
}

// applyMessageFlagsSet sets the flags of the given messages.
func (state *State) applyMessageFlagsSet(ctx context.Context, tx *ent.Tx, messageIDs []imap.InternalMessageID, setFlags imap.FlagSet) error {
	if setFlags.Contains(imap.FlagRecent) {
		panic("the recent flag is read-only")
	}

	if err := DBSetDeletedFlag(ctx, tx, state.snap.mboxID.InternalID, messageIDs, setFlags.Contains(imap.FlagDeleted)); err != nil {
		return err
	}

	if err := DBSetMessageFlags(ctx, tx, messageIDs, setFlags.Remove(imap.FlagDeleted)); err != nil {
		return err
	}

	for _, messageID := range messageIDs {
		if err := state.forStateWithMessage(messageID, func(other *State) error {
			newFlags := setFlags

			if other.snap.mboxID != state.snap.mboxID {
				snapFlags, err := other.snap.getMessageFlags(messageID)
				if err != nil {
					return err
				}

				newFlags = newFlags.Set(imap.FlagDeleted, snapFlags.Contains(imap.FlagDeleted))
			}

			return other.pushResponder(ctx, tx, newFetch(
				messageID,
				newFlags,
				isUID(ctx),
				other == state && isSilent(ctx),
			))
		}); err != nil {
			return err
		}
	}

	return nil
}
