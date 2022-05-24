package backend

import (
	"context"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/backend/ent"
	"github.com/bradenaw/juniper/xslices"
)

// applyMessageFlagsAdded adds the flags to the given messages.
func (state *State) applyMessageFlagsAdded(ctx context.Context, tx *ent.Tx, messageIDs []string, addFlags imap.FlagSet) error {
	if addFlags.Contains(imap.FlagRecent) {
		panic("the recent flag is read-only")
	}

	curFlags, err := txGetMessageFlags(ctx, tx, messageIDs)
	if err != nil {
		return err
	}

	delFlags, err := txGetMessageDeleted(ctx, tx, state.snap.mboxID, messageIDs)
	if err != nil {
		return err
	}

	if addFlags.Contains(imap.FlagDeleted) {
		if err := txSetDeletedFlag(ctx, tx, state.snap.mboxID, xslices.Filter(messageIDs, func(messageID string) bool {
			return !delFlags[messageID]
		}), true); err != nil {
			return err
		}
	}

	for _, flag := range addFlags.Remove(imap.FlagDeleted).ToSlice() {
		if err := txAddMessageFlag(ctx, tx, xslices.Filter(messageIDs, func(messageID string) bool {
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
func (state *State) applyMessageFlagsRemoved(ctx context.Context, tx *ent.Tx, messageIDs []string, remFlags imap.FlagSet) error {
	if remFlags.Contains(imap.FlagRecent) {
		panic("the recent flag is read-only")
	}

	curFlags, err := txGetMessageFlags(ctx, tx, messageIDs)
	if err != nil {
		return err
	}

	delFlags, err := txGetMessageDeleted(ctx, tx, state.snap.mboxID, messageIDs)
	if err != nil {
		return err
	}

	if remFlags.Contains(imap.FlagDeleted) {
		if err := txSetDeletedFlag(ctx, tx, state.snap.mboxID, xslices.Filter(messageIDs, func(messageID string) bool {
			return delFlags[messageID]
		}), false); err != nil {
			return err
		}
	}

	for _, flag := range remFlags.Remove(imap.FlagDeleted).ToSlice() {
		if err := txRemoveMessageFlag(ctx, tx, xslices.Filter(messageIDs, func(messageID string) bool {
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

	res := make(map[string]imap.FlagSet)

	for _, messageID := range messageIDs {
		res[messageID] = curFlags[messageID].RemoveFlagSet(remFlags)
	}

	return nil
}

// applyMessageFlagsSet sets the flags of the given messages.
func (state *State) applyMessageFlagsSet(ctx context.Context, tx *ent.Tx, messageIDs []string, setFlags imap.FlagSet) error {
	if setFlags.Contains(imap.FlagRecent) {
		panic("the recent flag is read-only")
	}

	if err := txSetDeletedFlag(ctx, tx, state.snap.mboxID, messageIDs, setFlags.Contains(imap.FlagDeleted)); err != nil {
		return err
	}

	if err := txSetMessageFlags(ctx, tx, messageIDs, setFlags.Remove(imap.FlagDeleted)); err != nil {
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
