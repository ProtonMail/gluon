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
	internalID, res, err := state.user.GetRemote().CreateMailbox(ctx, strings.Split(name, state.delimiter))
	if err != nil {
		return nil, err
	}

	exists, err := db.MailboxExistsWithID(ctx, tx.Client(), internalID)
	if err != nil {
		return nil, err
	}

	if !exists {
		mbox, err := db.CreateMailbox(
			ctx,
			tx,
			internalID,
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

	return db.GetMailboxByID(ctx, tx.Client(), internalID)
}

func (state *State) actionCreateMailbox(ctx context.Context, tx *ent.Tx, name string) error {
	internalID, res, err := state.user.GetRemote().CreateMailbox(ctx, strings.Split(name, state.delimiter))
	if err != nil {
		return err
	}

	return db.CreateMailboxIfNotExists(ctx, tx, internalID, res, state.delimiter)
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
) (int, error) {
	internalID, res, err := state.user.GetRemote().CreateMessage(ctx, mboxID.RemoteID, literal, flags, date)
	if err != nil {
		return 0, err
	}

	update := imap.NewMessagesCreated()

	if err := update.Add(res, literal, mboxID.RemoteID); err != nil {
		return 0, err
	}

	var reqs []*db.CreateMessageReq

	{
		msg := update.Messages[0]
		literal, err := rfc822.SetHeaderValue(msg.Literal, ids.InternalIDKey, string(internalID))
		if err != nil {
			return 0, fmt.Errorf("failed to set internal ID: %w", err)
		}

		if err := stx.Set(internalID, literal); err != nil {
			return 0, fmt.Errorf("failed to store message literal: %w", err)
		}

		reqs = append(reqs, &db.CreateMessageReq{
			Message:    msg.Message,
			Literal:    literal,
			Body:       msg.Body,
			Structure:  msg.Structure,
			Envelope:   msg.Envelope,
			InternalID: internalID,
		})
	}

	if _, err := db.CreateMessages(ctx, tx, reqs...); err != nil {
		return 0, fmt.Errorf("failed to create message: %w", err)
	}

	msgIDs := []imap.InternalMessageID{internalID}

	messageUIDs, err := db.AddMessagesToMailbox(ctx, tx, msgIDs, mboxID.InternalID)
	if err != nil {
		return 0, err
	}

	// We can append to non-selected mailboxes.
	var st *State
	if isSelectedMailbox {
		st = state
	}

	uid := messageUIDs[internalID]
	if err := state.user.QueueOrApplyStateUpdate(
		ctx,
		tx,
		newExistsStateUpdateWithExists(
			mboxID.InternalID,
			[]*exists{newExists(ids.MessageIDPair{InternalID: internalID, RemoteID: res.ID}, uid.UID, db.NewFlagSet(uid, uid.Edges.Message.Edges.Flags))},
			st,
		),
	); err != nil {
		return 0, err
	}

	return messageUIDs[internalID].UID, nil
}

func (state *State) actionAddMessagesToMailbox(
	ctx context.Context,
	tx *ent.Tx,
	messageIDs []ids.MessageIDPair,
	mboxID ids.MailboxIDPair,
	isMailboxSelected bool,
) (map[imap.InternalMessageID]*ent.UID, error) {
	var haveMessageIDs []ids.MessageIDPair
	if state.snap != nil && state.snap.mboxID.InternalID == mboxID.InternalID {
		haveMessageIDs = state.snap.getAllMessageIDs()
	} else {
		msgs, err := db.GetMailboxMessageIDPairs(ctx, tx.Client(), mboxID.InternalID)
		if err != nil {
			return nil, err
		}

		haveMessageIDs = msgs
	}

	if remMessageIDs := xslices.Filter(messageIDs, func(messageID ids.MessageIDPair) bool {
		return slices.Contains(haveMessageIDs, messageID)
	}); len(remMessageIDs) > 0 {
		if err := state.actionRemoveMessagesFromMailbox(ctx, tx, remMessageIDs, mboxID); err != nil {
			return nil, err
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

func (state *State) actionRemoveMessagesFromMailbox(
	ctx context.Context,
	tx *ent.Tx,
	messageIDs []ids.MessageIDPair,
	mboxID ids.MailboxIDPair,
) error {
	haveMessageIDs, err := db.GetMailboxMessageIDPairs(ctx, tx.Client(), mboxID.InternalID)
	if err != nil {
		return err
	}

	messageIDs = xslices.Filter(messageIDs, func(messageID ids.MessageIDPair) bool {
		return slices.Contains(haveMessageIDs, messageID)
	})

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

func (state *State) actionMoveMessages(
	ctx context.Context,
	tx *ent.Tx,
	messageIDs []ids.MessageIDPair,
	mboxFromID, mboxToID ids.MailboxIDPair,
) (map[imap.InternalMessageID]*ent.UID, error) {
	if mboxFromID.InternalID == mboxToID.InternalID {
		internalIDs, _ := ids.SplitMessageIDPairSlice(messageIDs)

		return db.BumpMailboxUIDsForMessage(ctx, tx, internalIDs, mboxToID.InternalID)
	}

	{
		var messageIDsToAdd []ids.MessageIDPair

		if state.snap != nil && state.snap.mboxID.InternalID == mboxToID.InternalID {
			messageIDsToAdd = state.snap.getAllMessageIDs()
		} else {
			msgs, err := db.GetMailboxMessageIDPairs(ctx, tx.Client(), mboxToID.InternalID)
			if err != nil {
				return nil, err
			}

			messageIDsToAdd = msgs
		}

		if remMessageIDs := xslices.Filter(messageIDs, func(messageID ids.MessageIDPair) bool {
			return slices.Contains(messageIDsToAdd, messageID)
		}); len(remMessageIDs) > 0 {
			if err := state.actionRemoveMessagesFromMailbox(ctx, tx, remMessageIDs, mboxToID); err != nil {
				return nil, err
			}
		}
	}

	messagesIDsToMove, err := db.GetMailboxMessageIDPairs(ctx, tx.Client(), mboxFromID.InternalID)
	if err != nil {
		return nil, err
	}

	messagesIDsToMove = xslices.Filter(messageIDs, func(messageID ids.MessageIDPair) bool {
		return slices.Contains(messagesIDsToMove, messageID)
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
	messageIDs []ids.MessageIDPair,
	addFlags imap.FlagSet,
) (map[imap.InternalMessageID]imap.FlagSet, error) {
	curFlags := make(map[imap.MessageID]imap.FlagSet)

	// Get the current flags that each message has.
	for _, messageID := range messageIDs {
		flags, err := state.snap.getMessageFlags(messageID.InternalID)
		if err != nil {
			return nil, err
		}

		curFlags[messageID.RemoteID] = flags
	}

	internalMsgIDs, remoteMsgIDs := ids.SplitMessageIDPairSlice(messageIDs)

	// If setting messages as seen, only set those messages that aren't currently seen.
	if addFlags.Contains(imap.FlagSeen) {
		if err := state.user.GetRemote().SetMessagesSeen(ctx, xslices.Filter(remoteMsgIDs, func(messageID imap.MessageID) bool {
			return !curFlags[messageID].Contains(imap.FlagSeen)
		}), true); err != nil {
			return nil, err
		}
	}

	// If setting messages as flagged, only set those messages that aren't currently flagged.
	if addFlags.Contains(imap.FlagFlagged) {
		if err := state.user.GetRemote().SetMessagesFlagged(ctx, xslices.Filter(remoteMsgIDs, func(messageID imap.MessageID) bool {
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

func (state *State) actionRemoveMessageFlags(
	ctx context.Context,
	tx *ent.Tx,
	messageIDs []ids.MessageIDPair,
	remFlags imap.FlagSet,
) (map[imap.InternalMessageID]imap.FlagSet, error) {
	curFlags := make(map[imap.MessageID]imap.FlagSet)

	// Get the current flags that each message has.
	for _, messageID := range messageIDs {
		flags, err := state.snap.getMessageFlags(messageID.InternalID)
		if err != nil {
			return nil, err
		}

		curFlags[messageID.RemoteID] = flags
	}

	internalMsgIDs, remoteMsgIDs := ids.SplitMessageIDPairSlice(messageIDs)

	// If setting messages as unseen, only set those messages that are currently seen.
	if remFlags.Contains(imap.FlagSeen) {
		if err := state.user.GetRemote().SetMessagesSeen(ctx, xslices.Filter(remoteMsgIDs, func(messageID imap.MessageID) bool {
			return curFlags[messageID].Contains(imap.FlagSeen)
		}), false); err != nil {
			return nil, err
		}
	}

	// If setting messages as unflagged, only set those messages that are currently flagged.
	if remFlags.Contains(imap.FlagFlagged) {
		if err := state.user.GetRemote().SetMessagesFlagged(ctx, xslices.Filter(remoteMsgIDs, func(messageID imap.MessageID) bool {
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

func (state *State) actionSetMessageFlags(ctx context.Context, tx *ent.Tx, messageIDs []ids.MessageIDPair, setFlags imap.FlagSet) error {
	if setFlags.Contains(imap.FlagRecent) {
		return fmt.Errorf("recent flag is read-only")
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

	internalMsgIDs, remoteMessageIDs := ids.SplitMessageIDPairSlice(messageIDs)

	// If setting messages as seen, only set those messages that aren't currently seen.
	if setFlags.Contains(imap.FlagSeen) {
		if err := state.user.GetRemote().SetMessagesSeen(ctx, xslices.Filter(remoteMessageIDs, func(messageID imap.MessageID) bool {
			return !curFlags[messageID].Contains(imap.FlagSeen)
		}), true); err != nil {
			return err
		}
	} else {
		if err := state.user.GetRemote().SetMessagesSeen(ctx, xslices.Filter(remoteMessageIDs, func(messageID imap.MessageID) bool {
			return curFlags[messageID].Contains(imap.FlagSeen)
		}), false); err != nil {
			return err
		}
	}

	// If setting messages as flagged, only set those messages that aren't currently flagged.
	if setFlags.Contains(imap.FlagFlagged) {
		if err := state.user.GetRemote().SetMessagesFlagged(ctx, xslices.Filter(remoteMessageIDs, func(messageID imap.MessageID) bool {
			return !curFlags[messageID].Contains(imap.FlagFlagged)
		}), true); err != nil {
			return err
		}
	} else {
		if err := state.user.GetRemote().SetMessagesFlagged(ctx, xslices.Filter(remoteMessageIDs, func(messageID imap.MessageID) bool {
			return curFlags[messageID].Contains(imap.FlagFlagged)
		}), false); err != nil {
			return err
		}
	}

	return state.applyMessageFlagsSet(ctx, tx, internalMsgIDs, setFlags)
}
