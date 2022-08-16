package backend

import (
	"context"
	"strings"
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/backend/ent"
	"github.com/bradenaw/juniper/xslices"
	"golang.org/x/exp/slices"
)

func (state *State) actionCreateMailbox(ctx context.Context, tx *ent.Tx, name string) (*ent.Mailbox, error) {
	internalID, res, err := state.remote.CreateMailbox(state.metadataID, strings.Split(name, state.delimiter))
	if err != nil {
		return nil, err
	}

	if err := state.apply(ctx, tx, imap.NewMailboxCreated(res)); err != nil {
		return nil, err
	}

	return DBGetMailboxByID(ctx, tx.Client(), internalID)
}

// TODO(REFACTOR): What if another client is selected in the same mailbox -- should we send expunge updates?
func (state *State) actionDeleteMailbox(ctx context.Context, tx *ent.Tx, mboxID imap.LabelID, name string) error {
	if err := state.remote.DeleteMailbox(state.metadataID, mboxID, strings.Split(name, state.delimiter)); err != nil {
		return err
	}

	return state.apply(ctx, tx, imap.NewMailboxDeleted(mboxID))
}

func (state *State) actionUpdateMailbox(ctx context.Context, tx *ent.Tx, mboxID imap.LabelID, oldName, newName string) error {
	if err := state.remote.UpdateMailbox(
		state.metadataID,
		mboxID,
		strings.Split(oldName, state.delimiter),
		strings.Split(newName, state.delimiter),
	); err != nil {
		return err
	}

	return state.apply(ctx, tx, imap.NewMailboxUpdated(mboxID, strings.Split(newName, state.delimiter)))
}

func (state *State) actionCreateMessage(ctx context.Context, tx *ent.Tx, mboxID MailboxIDPair, literal []byte, flags imap.FlagSet, date time.Time) (int, error) {
	internalID, res, err := state.remote.CreateMessage(state.metadataID, mboxID.RemoteID, literal, flags, date)
	if err != nil {
		return 0, err
	}

	update := imap.NewMessagesCreated()

	if err := update.Add(res, literal, []imap.LabelID{mboxID.RemoteID}); err != nil {
		return 0, err
	}

	if err := state.apply(ctx, tx, update); err != nil {
		return 0, err
	}

	messageUIDs, err := DBGetMessageUIDs(ctx, tx.Client(), mboxID.InternalID, []imap.InternalMessageID{internalID})
	if err != nil {
		return 0, err
	}

	return messageUIDs[internalID], nil
}

func (state *State) actionAddMessagesToMailbox(ctx context.Context,
	tx *ent.Tx,
	messageIDs []MessageIDPair,
	mboxID MailboxIDPair,
) (map[imap.InternalMessageID]int, error) {
	var haveMessageIDs []MessageIDPair

	if state.snap != nil && state.snap.mboxID.InternalID == mboxID.InternalID {
		haveMessageIDs = state.snap.getAllMessageIDs()
	} else {
		var err error

		if haveMessageIDs, err = DBGetMailboxMessageIDPairs(ctx, tx.Client(), mboxID.InternalID); err != nil {
			return nil, err
		}
	}

	if remMessageIDs := xslices.Filter(messageIDs, func(messageID MessageIDPair) bool {
		return slices.Contains(haveMessageIDs, messageID)
	}); len(remMessageIDs) > 0 {
		if err := state.actionRemoveMessagesFromMailbox(ctx, tx, remMessageIDs, mboxID); err != nil {
			return nil, err
		}
	}

	internalIDs, remoteIDs := SplitMessageIDPairSlice(messageIDs)

	if err := state.remote.AddMessagesToMailbox(state.metadataID, remoteIDs, mboxID.RemoteID); err != nil {
		return nil, err
	}

	return state.applyMessagesAddedToMailbox(ctx, tx, mboxID.InternalID, internalIDs)
}

func (state *State) actionRemoveMessagesFromMailbox(ctx context.Context, tx *ent.Tx, messageIDs []MessageIDPair, mboxID MailboxIDPair) error {
	haveMessageIDs, err := DBGetMailboxMessageIDPairs(ctx, tx.Client(), mboxID.InternalID)
	if err != nil {
		return err
	}

	messageIDs = xslices.Filter(messageIDs, func(messageID MessageIDPair) bool {
		return slices.Contains(haveMessageIDs, messageID)
	})

	internalIDs, remoteIDs := SplitMessageIDPairSlice(messageIDs)

	if err := state.remote.RemoveMessagesFromMailbox(state.metadataID, remoteIDs, mboxID.RemoteID); err != nil {
		return err
	}

	return state.applyMessagesRemovedFromMailbox(ctx, tx, mboxID.InternalID, internalIDs)
}

func (state *State) actionAddMessageFlags(ctx context.Context, tx *ent.Tx, messageIDs []MessageIDPair, addFlags imap.FlagSet) (map[imap.InternalMessageID]imap.FlagSet, error) {
	curFlags := make(map[imap.MessageID]imap.FlagSet)

	// Get the current flags that each message has.
	for _, messageID := range messageIDs {
		flags, err := state.snap.getMessageFlags(messageID.InternalID)
		if err != nil {
			return nil, err
		}

		curFlags[messageID.RemoteID] = flags
	}

	internalMsgIDs, remoteMsgIDs := SplitMessageIDPairSlice(messageIDs)

	// If setting messages as seen, only set those messages that aren't currently seen.
	if addFlags.Contains(imap.FlagSeen) {
		if err := state.remote.SetMessagesSeen(state.metadataID, xslices.Filter(remoteMsgIDs, func(messageID imap.MessageID) bool {
			return !curFlags[messageID].Contains(imap.FlagSeen)
		}), true); err != nil {
			return nil, err
		}
	}

	// If setting messages as flagged, only set those messages that aren't currently flagged.
	if addFlags.Contains(imap.FlagFlagged) {
		if err := state.remote.SetMessagesFlagged(state.metadataID, xslices.Filter(remoteMsgIDs, func(messageID imap.MessageID) bool {
			return !curFlags[messageID].Contains(imap.FlagFlagged)
		}), true); err != nil {
			return nil, err
		}
	}

	if err := state.applyMessageFlagsAdded(ctx, tx, internalMsgIDs, addFlags); err != nil {
		return nil, err
	}

	res := make(map[imap.InternalMessageID]imap.FlagSet)

	for _, messageID := range messageIDs {
		res[messageID.InternalID] = curFlags[messageID.RemoteID].AddFlagSet(addFlags)
	}

	return res, nil
}

func (state *State) actionRemoveMessageFlags(ctx context.Context, tx *ent.Tx, messageIDs []MessageIDPair, remFlags imap.FlagSet) (map[imap.InternalMessageID]imap.FlagSet, error) {
	curFlags := make(map[imap.MessageID]imap.FlagSet)

	// Get the current flags that each message has.
	for _, messageID := range messageIDs {
		flags, err := state.snap.getMessageFlags(messageID.InternalID)
		if err != nil {
			return nil, err
		}

		curFlags[messageID.RemoteID] = flags
	}

	internalMsgIDs, remoteMsgIDs := SplitMessageIDPairSlice(messageIDs)

	// If setting messages as unseen, only set those messages that are currently seen.
	if remFlags.Contains(imap.FlagSeen) {
		if err := state.remote.SetMessagesSeen(state.metadataID, xslices.Filter(remoteMsgIDs, func(messageID imap.MessageID) bool {
			return curFlags[messageID].Contains(imap.FlagSeen)
		}), false); err != nil {
			return nil, err
		}
	}

	// If setting messages as unflagged, only set those messages that are currently flagged.
	if remFlags.Contains(imap.FlagFlagged) {
		if err := state.remote.SetMessagesFlagged(state.metadataID, xslices.Filter(remoteMsgIDs, func(messageID imap.MessageID) bool {
			return curFlags[messageID].Contains(imap.FlagFlagged)
		}), false); err != nil {
			return nil, err
		}
	}

	if err := state.applyMessageFlagsRemoved(ctx, tx, internalMsgIDs, remFlags); err != nil {
		return nil, err
	}

	res := make(map[imap.InternalMessageID]imap.FlagSet)

	for _, messageID := range messageIDs {
		res[messageID.InternalID] = curFlags[messageID.RemoteID].RemoveFlagSet(remFlags)
	}

	return res, nil
}

func (state *State) actionSetMessageFlags(ctx context.Context, tx *ent.Tx, messageIDs []MessageIDPair, setFlags imap.FlagSet) error {
	if setFlags.Contains(imap.FlagRecent) {
		panic("recent flag is read-only")
	}

	curFlags := make(map[imap.MessageID]imap.FlagSet)

	// Get the current flags that each message has.
	for _, messageID := range messageIDs {
		flags, err := state.snap.getMessageFlags(messageID.InternalID)
		if err != nil {
			return err
		}

		curFlags[messageID.RemoteID] = flags
	}

	internalMsgIDs, remoteMessageIDs := SplitMessageIDPairSlice(messageIDs)

	// If setting messages as seen, only set those messages that aren't currently seen.
	if setFlags.Contains(imap.FlagSeen) {
		if err := state.remote.SetMessagesSeen(state.metadataID, xslices.Filter(remoteMessageIDs, func(messageID imap.MessageID) bool {
			return !curFlags[messageID].Contains(imap.FlagSeen)
		}), true); err != nil {
			return err
		}
	} else {
		if err := state.remote.SetMessagesSeen(state.metadataID, xslices.Filter(remoteMessageIDs, func(messageID imap.MessageID) bool {
			return curFlags[messageID].Contains(imap.FlagSeen)
		}), false); err != nil {
			return err
		}
	}

	// If setting messages as flagged, only set those messages that aren't currently flagged.
	if setFlags.Contains(imap.FlagFlagged) {
		if err := state.remote.SetMessagesFlagged(state.metadataID, xslices.Filter(remoteMessageIDs, func(messageID imap.MessageID) bool {
			return !curFlags[messageID].Contains(imap.FlagFlagged)
		}), true); err != nil {
			return err
		}
	} else {
		if err := state.remote.SetMessagesFlagged(state.metadataID, xslices.Filter(remoteMessageIDs, func(messageID imap.MessageID) bool {
			return curFlags[messageID].Contains(imap.FlagFlagged)
		}), false); err != nil {
			return err
		}
	}

	return state.applyMessageFlagsSet(ctx, tx, internalMsgIDs, setFlags)
}
