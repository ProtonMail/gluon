package backend

import (
	"context"
	"strings"
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/backend/ent"
	"github.com/ProtonMail/gluon/internal/backend/ent/mailbox"
	"github.com/bradenaw/juniper/xslices"
	"golang.org/x/exp/slices"
)

func (state *State) actionCreateMailbox(ctx context.Context, tx *ent.Tx, name string) (*ent.Mailbox, error) {
	res, err := state.remote.CreateMailbox(ctx, strings.Split(name, state.delimiter))
	if err != nil {
		return nil, err
	}

	if err := state.apply(ctx, tx, imap.NewMailboxCreated(res)); err != nil {
		return nil, err
	}

	return tx.Mailbox.Query().Where(mailbox.MailboxID(res.ID)).Only(ctx)
}

// TODO(REFACTOR): What if another client is selected in the same mailbox -- should we send expunge updates?
func (state *State) actionDeleteMailbox(ctx context.Context, tx *ent.Tx, mboxID, name string) error {
	if err := state.remote.DeleteMailbox(ctx, mboxID, strings.Split(name, state.delimiter)); err != nil {
		return err
	}

	return state.apply(ctx, tx, imap.NewMailboxDeleted(mboxID))
}

func (state *State) actionUpdateMailbox(ctx context.Context, tx *ent.Tx, mboxID, oldName, newName string) error {
	if err := state.remote.UpdateMailbox(
		ctx,
		mboxID,
		strings.Split(oldName, state.delimiter),
		strings.Split(newName, state.delimiter),
	); err != nil {
		return err
	}

	return state.apply(ctx, tx, imap.NewMailboxUpdated(mboxID, strings.Split(newName, state.delimiter)))
}

func (state *State) actionCreateMessage(ctx context.Context, tx *ent.Tx, mboxID string, literal []byte, flags imap.FlagSet, date time.Time) (int, error) {
	res, err := state.remote.CreateMessage(ctx, mboxID, literal, flags, date)
	if err != nil {
		return 0, err
	}

	update, err := imap.NewMessageCreated(res, literal, []string{mboxID})
	if err != nil {
		return 0, err
	}

	if err := state.apply(ctx, tx, update); err != nil {
		return 0, err
	}

	messageUIDs, err := txGetMessageUIDs(ctx, tx, mboxID, []string{res.ID})
	if err != nil {
		return 0, err
	}

	return messageUIDs[res.ID], nil
}

func (state *State) actionAddMessagesToMailbox(ctx context.Context, tx *ent.Tx, messageIDs []string, mboxID string) (map[string]int, error) {
	var haveMessageIDs []string

	if state.snap != nil && state.snap.mboxID == mboxID {
		haveMessageIDs = state.snap.getAllMessageIDs()
	} else {
		var err error

		if haveMessageIDs, err = txGetMailboxMessageIDs(ctx, tx, mboxID); err != nil {
			return nil, err
		}
	}

	if remMessageIDs := xslices.Filter(messageIDs, func(messageID string) bool {
		return slices.Contains(haveMessageIDs, messageID)
	}); len(remMessageIDs) > 0 {
		if err := state.actionRemoveMessagesFromMailbox(ctx, tx, remMessageIDs, mboxID); err != nil {
			return nil, err
		}
	}

	if err := state.remote.AddMessagesToMailbox(ctx, messageIDs, mboxID); err != nil {
		return nil, err
	}

	return state.applyMessagesAddedToMailbox(ctx, tx, mboxID, messageIDs)
}

func (state *State) actionRemoveMessagesFromMailbox(ctx context.Context, tx *ent.Tx, messageIDs []string, mboxID string) error {
	var haveMessageIDs []string

	if state.snap != nil && state.snap.mboxID == mboxID {
		haveMessageIDs = xslices.Filter(state.snap.getAllMessageIDs(), func(messageID string) bool {
			return !state.pool.hasMessage(mboxID, messageID)
		})
	} else {
		var err error

		if haveMessageIDs, err = txGetMailboxMessageIDs(ctx, tx, mboxID); err != nil {
			return err
		}
	}

	messageIDs = xslices.Filter(messageIDs, func(messageID string) bool {
		return slices.Contains(haveMessageIDs, messageID)
	})

	if err := state.remote.RemoveMessagesFromMailbox(ctx, messageIDs, mboxID); err != nil {
		return err
	}

	return state.applyMessagesRemovedFromMailbox(ctx, tx, mboxID, messageIDs)
}

func (state *State) actionAddMessageFlags(ctx context.Context, tx *ent.Tx, messageIDs []string, addFlags imap.FlagSet) (map[string]imap.FlagSet, error) {
	curFlags := make(map[string]imap.FlagSet)

	// Get the current flags that each message has.
	for _, messageID := range messageIDs {
		flags, err := state.snap.getMessageFlags(messageID)
		if err != nil {
			return nil, err
		}

		curFlags[messageID] = flags
	}

	// If setting messages as seen, only set those messages that aren't currently seen.
	if addFlags.Contains(imap.FlagSeen) {
		if err := state.remote.SetMessagesSeen(ctx, xslices.Filter(messageIDs, func(messageID string) bool {
			return !curFlags[messageID].Contains(imap.FlagSeen)
		}), true); err != nil {
			return nil, err
		}
	}

	// If setting messages as flagged, only set those messages that aren't currently flagged.
	if addFlags.Contains(imap.FlagFlagged) {
		if err := state.remote.SetMessagesFlagged(ctx, xslices.Filter(messageIDs, func(messageID string) bool {
			return !curFlags[messageID].Contains(imap.FlagFlagged)
		}), true); err != nil {
			return nil, err
		}
	}

	if err := state.applyMessageFlagsAdded(ctx, tx, messageIDs, addFlags); err != nil {
		return nil, err
	}

	res := make(map[string]imap.FlagSet)

	for _, messageID := range messageIDs {
		res[messageID] = curFlags[messageID].AddFlagSet(addFlags)
	}

	return res, nil
}

func (state *State) actionRemoveMessageFlags(ctx context.Context, tx *ent.Tx, messageIDs []string, remFlags imap.FlagSet) (map[string]imap.FlagSet, error) {
	curFlags := make(map[string]imap.FlagSet)

	// Get the current flags that each message has.
	for _, messageID := range messageIDs {
		flags, err := state.snap.getMessageFlags(messageID)
		if err != nil {
			return nil, err
		}

		curFlags[messageID] = flags
	}

	// If setting messages as unseen, only set those messages that are currently seen.
	if remFlags.Contains(imap.FlagSeen) {
		if err := state.remote.SetMessagesSeen(ctx, xslices.Filter(messageIDs, func(messageID string) bool {
			return curFlags[messageID].Contains(imap.FlagSeen)
		}), false); err != nil {
			return nil, err
		}
	}

	// If setting messages as unflagged, only set those messages that are currently flagged.
	if remFlags.Contains(imap.FlagFlagged) {
		if err := state.remote.SetMessagesFlagged(ctx, xslices.Filter(messageIDs, func(messageID string) bool {
			return curFlags[messageID].Contains(imap.FlagFlagged)
		}), false); err != nil {
			return nil, err
		}
	}

	if err := state.applyMessageFlagsRemoved(ctx, tx, messageIDs, remFlags); err != nil {
		return nil, err
	}

	res := make(map[string]imap.FlagSet)

	for _, messageID := range messageIDs {
		res[messageID] = curFlags[messageID].RemoveFlagSet(remFlags)
	}

	return res, nil
}

func (state *State) actionSetMessageFlags(ctx context.Context, tx *ent.Tx, messageIDs []string, setFlags imap.FlagSet) error {
	if setFlags.Contains(imap.FlagRecent) {
		panic("recent flag is read-only")
	}

	curFlags := make(map[string]imap.FlagSet)

	// Get the current flags that each message has.
	for _, messageID := range messageIDs {
		flags, err := state.snap.getMessageFlags(messageID)
		if err != nil {
			return err
		}

		curFlags[messageID] = flags
	}

	// If setting messages as seen, only set those messages that aren't currently seen.
	if setFlags.Contains(imap.FlagSeen) {
		if err := state.remote.SetMessagesSeen(ctx, xslices.Filter(messageIDs, func(messageID string) bool {
			return !curFlags[messageID].Contains(imap.FlagSeen)
		}), true); err != nil {
			return err
		}
	} else {
		if err := state.remote.SetMessagesSeen(ctx, xslices.Filter(messageIDs, func(messageID string) bool {
			return curFlags[messageID].Contains(imap.FlagSeen)
		}), false); err != nil {
			return err
		}
	}

	// If setting messages as flagged, only set those messages that aren't currently flagged.
	if setFlags.Contains(imap.FlagFlagged) {
		if err := state.remote.SetMessagesFlagged(ctx, xslices.Filter(messageIDs, func(messageID string) bool {
			return !curFlags[messageID].Contains(imap.FlagFlagged)
		}), true); err != nil {
			return err
		}
	} else {
		if err := state.remote.SetMessagesFlagged(ctx, xslices.Filter(messageIDs, func(messageID string) bool {
			return curFlags[messageID].Contains(imap.FlagFlagged)
		}), false); err != nil {
			return err
		}
	}

	return state.applyMessageFlagsSet(ctx, tx, messageIDs, setFlags)
}
