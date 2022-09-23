package state

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db"
	"github.com/ProtonMail/gluon/internal/db/ent"
	"github.com/ProtonMail/gluon/internal/ids"
	"github.com/ProtonMail/gluon/rfc822"
	"github.com/ProtonMail/gluon/store"
	"github.com/bradenaw/juniper/xslices"
	"golang.org/x/exp/slices"
)

func (state *State) actionCreateAndGetMailbox(ctx context.Context, tx *ent.Tx, name string) (*ent.Mailbox, error) {
	res, err := state.user.GetRemote().CreateMailbox(ctx, strings.Split(name, state.delimiter))
	if err != nil {
		return nil, err
	}

	exists, err := db.MailboxExistsWithRemoteID(ctx, tx.Client(), res.ID)
	if err != nil {
		return nil, err
	}

	if !exists {
		mbox, err := db.CreateMailbox(
			ctx,
			tx,
			res.ID,
			strings.Join(res.Name, state.user.GetDelimiter()),
			res.Flags,
			res.PermanentFlags,
			res.Attributes,
		)
		if err != nil {
			return nil, err
		}

		return mbox, nil
	}

	return db.GetMailboxByRemoteID(ctx, tx.Client(), res.ID)
}

func (state *State) actionCreateMailbox(ctx context.Context, tx *ent.Tx, name string) error {
	res, err := state.user.GetRemote().CreateMailbox(ctx, strings.Split(name, state.delimiter))
	if err != nil {
		return err
	}

	return db.CreateMailboxIfNotExists(ctx, tx, res, state.delimiter)
}

func (state *State) actionDeleteMailbox(ctx context.Context, tx *ent.Tx, mboxID ids.MailboxIDPair) error {
	if err := state.user.GetRemote().DeleteMailbox(ctx, mboxID.RemoteID); err != nil {
		return err
	}

	if err := db.DeleteMailboxWithRemoteID(ctx, tx, mboxID.RemoteID); err != nil {
		return err
	}

	return state.user.QueueOrApplyStateUpdate(ctx, tx, NewMailboxDeletedStateUpdate(mboxID.InternalID))
}

func (state *State) actionUpdateMailbox(ctx context.Context, tx *ent.Tx, mboxID imap.LabelID, oldName, newName string) error {
	if err := state.user.GetRemote().UpdateMailbox(
		ctx,
		mboxID,
		strings.Split(oldName, state.delimiter),
		strings.Split(newName, state.delimiter),
	); err != nil {
		return err
	}

	return db.RenameMailboxWithRemoteID(ctx, tx, mboxID, newName)
}

func (state *State) actionCreateMessage(
	ctx context.Context,
	tx *ent.Tx,
	stx store.Transaction,
	mboxID ids.MailboxIDPair,
	literal []byte,
	flags imap.FlagSet,
	date time.Time,
	isSelectedMailbox bool,
) (imap.UID, error) {
	parsedMessage, err := imap.NewParsedMessage(literal)
	if err != nil {
		return 0, err
	}

	internalID, res, err := state.user.GetRemote().CreateMessage(ctx, mboxID.RemoteID, literal, parsedMessage, flags, date)
	if err != nil {
		return 0, err
	}

	literalWithHeader, err := rfc822.SetHeaderValue(literal, ids.InternalIDKey, internalID.String())

	if err != nil {
		return 0, fmt.Errorf("failed to set internal ID: %w", err)
	}

	if err := stx.Set(internalID, literalWithHeader); err != nil {
		return 0, fmt.Errorf("failed to store message literal: %w", err)
	}

	req := db.CreateMessageReq{
		Message:    res,
		Literal:    literalWithHeader,
		Body:       parsedMessage.Body,
		Structure:  parsedMessage.Structure,
		Envelope:   parsedMessage.Envelope,
		InternalID: internalID,
	}

	messageUID, flagSet, err := db.CreateAndAddMessageToMailbox(ctx, tx, mboxID.InternalID, &req)
	if err != nil {
		return 0, err
	}

	// We can append to non-selected mailboxes.
	var st *State
	if isSelectedMailbox {
		st = state
	}

	if err := state.user.QueueOrApplyStateUpdate(
		ctx,
		tx,
		newExistsStateUpdateWithExists(
			mboxID.InternalID,
			[]*exists{newExists(ids.MessageIDPair{InternalID: internalID, RemoteID: res.ID}, messageUID, flagSet)},
			st,
		),
	); err != nil {
		return 0, err
	}

	return messageUID, nil
}

func (state *State) actionAddMessagesToMailbox(
	ctx context.Context,
	tx *ent.Tx,
	messageIDs []ids.MessageIDPair,
	mboxID ids.MailboxIDPair,
	isMailboxSelected bool,
) ([]db.UIDWithFlags, error) {
	{
		haveMessageIDs, err := db.FilterMailboxContains(ctx, tx.Client(), mboxID.InternalID, messageIDs)
		if err != nil {
			return nil, err
		}

		if remMessageIDs := xslices.Filter(messageIDs, func(messageID ids.MessageIDPair) bool {
			return slices.Contains(haveMessageIDs, messageID.InternalID)
		}); len(remMessageIDs) > 0 {
			if err := state.actionRemoveMessagesFromMailboxUnchecked(ctx, tx, remMessageIDs, mboxID); err != nil {
				return nil, err
			}
		}
	}

	internalIDs, remoteIDs := ids.SplitMessageIDPairSlice(messageIDs)

	if err := state.user.GetRemote().AddMessagesToMailbox(ctx, remoteIDs, mboxID.RemoteID); err != nil {
		return nil, err
	}

	// Messages can be added to a mailbox that is not selected.
	var st *State
	if isMailboxSelected {
		st = state
	}

	messageUIDs, update, err := AddMessagesToMailbox(ctx, tx, mboxID.InternalID, internalIDs, st)
	if err != nil {
		return nil, err
	}

	if err := state.user.QueueOrApplyStateUpdate(ctx, tx, update); err != nil {
		return nil, err
	}

	return messageUIDs, nil
}

// actionRemoveMessagesFromMailboxUnchecked is similar to actionRemoveMessagesFromMailbox, but it does not validate
// the input for whether messages actually exist in the database or if the message set is empty. use this when you
// have already validated the input beforehand (e.g.: actionAddMessagesToMailbox and actionRemoveMessagesFromMailbox).
func (state *State) actionRemoveMessagesFromMailboxUnchecked(
	ctx context.Context,
	tx *ent.Tx,
	messageIDs []ids.MessageIDPair,
	mboxID ids.MailboxIDPair,
) error {
	internalIDs, remoteIDs := ids.SplitMessageIDPairSlice(messageIDs)

	if err := state.user.GetRemote().RemoveMessagesFromMailbox(ctx, remoteIDs, mboxID.RemoteID); err != nil {
		return err
	}

	updates, err := RemoveMessagesFromMailbox(ctx, tx, mboxID.InternalID, internalIDs)
	if err != nil {
		return err
	}

	for _, update := range updates {
		if err := state.user.QueueOrApplyStateUpdate(ctx, tx, update); err != nil {
			return err
		}
	}

	return nil
}

func (state *State) actionRemoveMessagesFromMailbox(
	ctx context.Context,
	tx *ent.Tx,
	messageIDs []ids.MessageIDPair,
	mboxID ids.MailboxIDPair,
) error {
	haveMessageIDs, err := db.FilterMailboxContains(ctx, tx.Client(), mboxID.InternalID, messageIDs)
	if err != nil {
		return err
	}

	messageIDs = xslices.Filter(messageIDs, func(messageID ids.MessageIDPair) bool {
		return slices.Contains(haveMessageIDs, messageID.InternalID)
	})

	if len(messageIDs) == 0 {
		return nil
	}

	return state.actionRemoveMessagesFromMailboxUnchecked(ctx, tx, messageIDs, mboxID)
}

func (state *State) actionMoveMessages(
	ctx context.Context,
	tx *ent.Tx,
	messageIDs []ids.MessageIDPair,
	mboxFromID, mboxToID ids.MailboxIDPair,
) ([]db.UIDWithFlags, error) {
	if mboxFromID.InternalID == mboxToID.InternalID {
		internalIDs, _ := ids.SplitMessageIDPairSlice(messageIDs)

		return db.BumpMailboxUIDsForMessage(ctx, tx, internalIDs, mboxToID.InternalID)
	}

	{
		messageIDsToAdd, err := db.FilterMailboxContains(ctx, tx.Client(), mboxToID.InternalID, messageIDs)
		if err != nil {
			return nil, err
		}

		if remMessageIDs := xslices.Filter(messageIDs, func(messageID ids.MessageIDPair) bool {
			return slices.Contains(messageIDsToAdd, messageID.InternalID)
		}); len(remMessageIDs) > 0 {
			if err := state.actionRemoveMessagesFromMailboxUnchecked(ctx, tx, remMessageIDs, mboxToID); err != nil {
				return nil, err
			}
		}
	}

	messageInFromMBox, err := db.FilterMailboxContains(ctx, tx.Client(), mboxFromID.InternalID, messageIDs)
	if err != nil {
		return nil, err
	}

	messagesIDsToMove := xslices.Filter(messageIDs, func(messageID ids.MessageIDPair) bool {
		return slices.Contains(messageInFromMBox, messageID.InternalID)
	})

	internalIDs, remoteIDs := ids.SplitMessageIDPairSlice(messagesIDsToMove)

	if err := state.user.GetRemote().MoveMessagesFromMailbox(ctx, remoteIDs, mboxFromID.RemoteID, mboxToID.RemoteID); err != nil {
		return nil, err
	}

	messageUIDs, updates, err := MoveMessagesFromMailbox(ctx, tx, mboxFromID.InternalID, mboxToID.InternalID, internalIDs, state)
	if err != nil {
		return nil, err
	}

	for _, update := range updates {
		if err := state.user.QueueOrApplyStateUpdate(ctx, tx, update); err != nil {
			return nil, err
		}
	}

	return messageUIDs, nil
}

func (state *State) actionAddMessageFlags(
	ctx context.Context,
	tx *ent.Tx,
	messages []*snapMsg,
	addFlags imap.FlagSet,
) error {
	internalMessageIDs := xslices.Map(messages, func(sm *snapMsg) imap.InternalMessageID {
		return sm.ID.InternalID
	})

	// If setting messages as seen, only set those messages that aren't currently seen.
	if addFlags.Contains(imap.FlagSeen) {
		var messagesToApply []imap.MessageID

		for _, msg := range messages {
			if !msg.flags.Contains(imap.FlagSeen) {
				messagesToApply = append(messagesToApply, msg.ID.RemoteID)
			}
		}

		if len(messagesToApply) != 0 {
			if err := state.user.GetRemote().SetMessagesSeen(ctx, messagesToApply, true); err != nil {
				return err
			}
		}
	}

	// If setting messages as flagged, only set those messages that aren't currently flagged.
	if addFlags.Contains(imap.FlagFlagged) {
		var messagesToApply []imap.MessageID

		for _, msg := range messages {
			if !msg.flags.Contains(imap.FlagFlagged) {
				messagesToApply = append(messagesToApply, msg.ID.RemoteID)
			}
		}

		if len(messagesToApply) != 0 {
			if err := state.user.GetRemote().SetMessagesFlagged(ctx, messagesToApply, true); err != nil {
				return err
			}
		}
	}

	if err := state.applyMessageFlagsAdded(ctx, tx, internalMessageIDs, addFlags); err != nil {
		return err
	}

	return nil
}

func (state *State) actionRemoveMessageFlags(
	ctx context.Context,
	tx *ent.Tx,
	messages []*snapMsg,
	remFlags imap.FlagSet,
) error {
	internalMessageIDs := xslices.Map(messages, func(sm *snapMsg) imap.InternalMessageID {
		return sm.ID.InternalID
	})

	// If setting messages as unseen, only set those messages that are currently seen.
	if remFlags.Contains(imap.FlagSeen) {
		var messagesToApply []imap.MessageID

		for _, msg := range messages {
			if msg.flags.Contains(imap.FlagSeen) {
				messagesToApply = append(messagesToApply, msg.ID.RemoteID)
			}
		}

		if len(messagesToApply) != 0 {
			if err := state.user.GetRemote().SetMessagesSeen(ctx, messagesToApply, false); err != nil {
				return err
			}
		}
	}

	// If setting messages as unflagged, only set those messages that are currently flagged.
	if remFlags.Contains(imap.FlagFlagged) {
		var messagesToApply []imap.MessageID

		for _, msg := range messages {
			if msg.flags.Contains(imap.FlagFlagged) {
				messagesToApply = append(messagesToApply, msg.ID.RemoteID)
			}
		}

		if len(messagesToApply) != 0 {
			if err := state.user.GetRemote().SetMessagesFlagged(ctx, messagesToApply, false); err != nil {
				return err
			}
		}
	}

	if err := state.applyMessageFlagsRemoved(ctx, tx, internalMessageIDs, remFlags); err != nil {
		return err
	}

	return nil
}

func (state *State) actionSetMessageFlags(ctx context.Context, tx *ent.Tx, messages []*snapMsg, setFlags imap.FlagSet) error {
	if setFlags.Contains(imap.FlagRecent) {
		return fmt.Errorf("recent flag is read-only")
	}

	internalMessageIDs := xslices.Map(messages, func(sm *snapMsg) imap.InternalMessageID {
		return sm.ID.InternalID
	})

	// If setting messages as seen, only set those messages that aren't currently seen.
	if setFlags.Contains(imap.FlagSeen) {
		var messagesToApply []imap.MessageID

		for _, msg := range messages {
			if !msg.flags.Contains(imap.FlagSeen) {
				messagesToApply = append(messagesToApply, msg.ID.RemoteID)
			}
		}

		if len(messagesToApply) != 0 {
			if err := state.user.GetRemote().SetMessagesSeen(ctx, messagesToApply, true); err != nil {
				return err
			}
		}
	} else {
		var messagesToApply []imap.MessageID

		for _, msg := range messages {
			if msg.flags.Contains(imap.FlagSeen) {
				messagesToApply = append(messagesToApply, msg.ID.RemoteID)
			}
		}

		if len(messagesToApply) != 0 {
			if err := state.user.GetRemote().SetMessagesSeen(ctx, messagesToApply, false); err != nil {
				return err
			}
		}
	}

	// If setting messages as flagged, only set those messages that aren't currently flagged.
	if setFlags.Contains(imap.FlagFlagged) {
		var messagesToApply []imap.MessageID

		for _, msg := range messages {
			if !msg.flags.Contains(imap.FlagFlagged) {
				messagesToApply = append(messagesToApply, msg.ID.RemoteID)
			}
		}

		if len(messagesToApply) != 0 {
			if err := state.user.GetRemote().SetMessagesFlagged(ctx, messagesToApply, true); err != nil {
				return err
			}
		}
	} else {
		var messagesToApply []imap.MessageID

		for _, msg := range messages {
			if msg.flags.Contains(imap.FlagFlagged) {
				messagesToApply = append(messagesToApply, msg.ID.RemoteID)
			}
		}

		if len(messagesToApply) != 0 {
			if err := state.user.GetRemote().SetMessagesFlagged(ctx, messagesToApply, false); err != nil {
				return err
			}
		}
	}

	return state.applyMessageFlagsSet(ctx, tx, internalMessageIDs, setFlags)
}
