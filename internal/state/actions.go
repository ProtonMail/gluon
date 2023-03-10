package state

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db"
	"github.com/ProtonMail/gluon/internal/db/ent"
	"github.com/ProtonMail/gluon/internal/ids"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/gluon/rfc822"
	"github.com/bradenaw/juniper/xslices"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

func (state *State) actionCreateAndGetMailbox(ctx context.Context, tx *ent.Tx, name string, uidValidity imap.UID) (*ent.Mailbox, error) {
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
			uidValidity,
		)
		if err != nil {
			return nil, err
		}

		return mbox, nil
	}

	return db.GetMailboxByRemoteID(ctx, tx.Client(), res.ID)
}

func (state *State) actionCreateMailbox(ctx context.Context, tx *ent.Tx, name string, uidValidity imap.UID) error {
	res, err := state.user.GetRemote().CreateMailbox(ctx, strings.Split(name, state.delimiter))
	if err != nil {
		return err
	}

	return db.CreateMailboxIfNotExists(ctx, tx, res, state.delimiter, uidValidity)
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

func (state *State) actionUpdateMailbox(ctx context.Context, tx *ent.Tx, mboxID imap.MailboxID, newName string) error {
	if err := state.user.GetRemote().UpdateMailbox(
		ctx,
		mboxID,
		strings.Split(newName, state.delimiter),
	); err != nil {
		return err
	}

	return db.RenameMailboxWithRemoteID(ctx, tx, mboxID, newName)
}

func (state *State) actionCreateMessage(
	ctx context.Context,
	tx *ent.Tx,
	mboxID ids.MailboxIDPair,
	literal []byte,
	flags imap.FlagSet,
	date time.Time,
	isSelectedMailbox bool,
	cameFromDrafts bool,
) (imap.UID, error) {
	internalID, res, newLiteral, err := state.user.GetRemote().CreateMessage(ctx, mboxID.RemoteID, literal, flags, date)
	if err != nil {
		return 0, err
	}

	{
		// Handle the case where duplicate messages can return the same remote ID.
		knownInternalID, knownErr := db.GetMessageIDFromRemoteID(ctx, tx.Client(), res.ID)
		if knownErr != nil && !ent.IsNotFound(knownErr) {
			return 0, knownErr
		}
		if knownErr == nil {
			// Try to collect the original message date.
			var existingMessageDate time.Time
			if existingMessage, msgErr := db.GetMessage(ctx, tx.Client(), internalID); msgErr == nil {
				existingMessageDate = existingMessage.Date
			}

			if cameFromDrafts {
				reporter.ExceptionWithContext(ctx, "Append to drafts must not return an existing RemoteID", reporter.Context{
					"remote-id":     res.ID,
					"new-date":      res.Date,
					"original-date": existingMessageDate,
				})

				logrus.Errorf("Append to drafts must not return an existing RemoteID (Remote=%v, Internal=%v)", res.ID, knownInternalID)

				return 0, fmt.Errorf("append to drafts returned an existing remote ID")
			}

			logrus.Debugf("Deduped message detected, adding existing %v message to mailbox instead.", knownInternalID.ShortID())

			result, err := state.actionAddMessagesToMailbox(ctx,
				tx,
				[]ids.MessageIDPair{{InternalID: knownInternalID, RemoteID: res.ID}},
				mboxID,
				isSelectedMailbox,
			)
			if err != nil {
				return 0, err
			}

			return result[0].UID, nil
		}
	}

	parsedMessage, err := imap.NewParsedMessage(newLiteral)
	if err != nil {
		return 0, err
	}

	literalWithHeader, literalSize, err := rfc822.SetHeaderValueNoMemCopy(newLiteral, ids.InternalIDKey, internalID.String())
	if err != nil {
		return 0, fmt.Errorf("failed to set internal ID: %w", err)
	}

	if err := state.user.GetStore().SetUnchecked(internalID, literalWithHeader); err != nil {
		return 0, fmt.Errorf("failed to store message literal: %w", err)
	}

	req := db.CreateMessageReq{
		Message:     res,
		LiteralSize: literalSize,
		Body:        parsedMessage.Body,
		Structure:   parsedMessage.Structure,
		Envelope:    parsedMessage.Envelope,
		InternalID:  internalID,
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

	if err := state.user.QueueOrApplyStateUpdate(ctx, tx, newExistsStateUpdateWithExists(
		mboxID.InternalID,
		[]*exists{newExists(ids.MessageIDPair{InternalID: internalID, RemoteID: res.ID}, messageUID, flagSet)},
		st,
	)); err != nil {
		return 0, err
	}

	return messageUID, nil
}

func (state *State) actionCreateRecoveredMessage(
	ctx context.Context,
	tx *ent.Tx,
	literal []byte,
	flags imap.FlagSet,
	date time.Time,
) error {
	internalID := imap.NewInternalMessageID()
	remoteID := ids.NewRecoveredRemoteMessageID(internalID)

	parsedMessage, err := imap.NewParsedMessage(literal)
	if err != nil {
		return err
	}

	if err := state.user.GetStore().SetUnchecked(internalID, bytes.NewReader(literal)); err != nil {
		return fmt.Errorf("failed to store message literal: %w", err)
	}

	req := db.CreateMessageReq{
		Message: imap.Message{
			ID:    remoteID,
			Flags: flags,
			Date:  date,
		},
		LiteralSize: len(literal),
		Body:        parsedMessage.Body,
		Structure:   parsedMessage.Structure,
		Envelope:    parsedMessage.Envelope,
		InternalID:  internalID,
	}

	recoveryMBoxID := state.user.GetRecoveryMailboxID()

	messageUID, flagSet, err := db.CreateAndAddMessageToMailbox(ctx, tx, recoveryMBoxID.InternalID, &req)
	if err != nil {
		return err
	}

	if err := state.user.QueueOrApplyStateUpdate(ctx, tx, newExistsStateUpdateWithExists(
		recoveryMBoxID.InternalID,
		[]*exists{newExists(ids.MessageIDPair{InternalID: internalID, RemoteID: remoteID}, messageUID, flagSet)},
		nil,
	)); err != nil {
		return err
	}

	return nil
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

	messageUIDs, update, err := AddMessagesToMailbox(ctx, tx, mboxID.InternalID, internalIDs, st, state.imapLimits)
	if err != nil {
		return nil, err
	}

	if err := state.user.QueueOrApplyStateUpdate(ctx, tx, update); err != nil {
		return nil, err
	}

	return messageUIDs, nil
}

func (state *State) actionAddRecoveredMessagesToMailbox(
	ctx context.Context,
	tx *ent.Tx,
	messageIDs []ids.MessageIDPair,
	mboxID ids.MailboxIDPair,
) ([]db.UIDWithFlags, Update, error) {
	internalIDs, remoteIDs := ids.SplitMessageIDPairSlice(messageIDs)

	if err := state.user.GetRemote().AddMessagesToMailbox(ctx, remoteIDs, mboxID.RemoteID); err != nil {
		return nil, nil, err
	}

	return AddMessagesToMailbox(ctx, tx, mboxID.InternalID, internalIDs, state, state.imapLimits)
}

func (state *State) actionImportRecoveredMessage(
	ctx context.Context,
	tx *ent.Tx,
	id imap.InternalMessageID,
	mboxID imap.MailboxID,
) (ids.MessageIDPair, bool, error) {
	message, err := db.GetImportedMessageData(ctx, tx.Client(), id)
	if err != nil {
		return ids.MessageIDPair{}, false, err
	}

	literal, err := state.user.GetStore().Get(id)
	if err != nil {
		return ids.MessageIDPair{}, false, err
	}

	messageFlags := imap.NewFlagSet()
	for _, flag := range message.Edges.Flags {
		messageFlags.AddToSelf(flag.Value)
	}

	internalID, res, newLiteral, err := state.user.GetRemote().CreateMessage(ctx, mboxID, literal, messageFlags, message.Date)
	if err != nil {
		return ids.MessageIDPair{}, false, err
	}

	{
		// Handle the unlikely case where duplicate messages can return the same remote ID.
		internalID, err := db.GetMessageIDFromRemoteID(ctx, tx.Client(), res.ID)
		if err != nil && !ent.IsNotFound(err) {
			return ids.MessageIDPair{}, false, err
		}

		if err == nil {
			return ids.MessageIDPair{
				InternalID: internalID,
				RemoteID:   res.ID,
			}, true, nil
		}
	}

	parsedMessage, err := imap.NewParsedMessage(newLiteral)
	if err != nil {
		return ids.MessageIDPair{}, false, err
	}

	literalReader, literalSize, err := rfc822.SetHeaderValueNoMemCopy(newLiteral, ids.InternalIDKey, internalID.String())
	if err != nil {
		return ids.MessageIDPair{}, false, fmt.Errorf("failed to set internal ID: %w", err)
	}

	if err := state.user.GetStore().SetUnchecked(internalID, literalReader); err != nil {
		return ids.MessageIDPair{}, false, fmt.Errorf("failed to store message literal: %w", err)
	}

	req := db.CreateMessageReq{
		Message:     res,
		LiteralSize: literalSize,
		Body:        parsedMessage.Body,
		Structure:   parsedMessage.Structure,
		Envelope:    parsedMessage.Envelope,
		InternalID:  internalID,
	}

	if _, err := db.CreateMessages(ctx, tx, &req); err != nil {
		return ids.MessageIDPair{}, false, err
	}

	return ids.MessageIDPair{
		InternalID: internalID,
		RemoteID:   res.ID,
	}, false, nil
}

func (state *State) actionCopyMessagesOutOfRecoveryMailbox(
	ctx context.Context,
	tx *ent.Tx,
	messageIDs []ids.MessageIDPair,
	mboxID ids.MailboxIDPair,
) ([]db.UIDWithFlags, error) {
	ids := make([]ids.MessageIDPair, 0, len(messageIDs))

	// Import messages to remote.
	for _, id := range messageIDs {
		id, _, err := state.actionImportRecoveredMessage(ctx, tx, id.InternalID, mboxID.RemoteID)
		if err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	// Label messages in destination.
	uidWithFlags, update, err := state.actionAddRecoveredMessagesToMailbox(ctx, tx, ids, mboxID)
	if err != nil {
		return nil, err
	}

	if err := state.user.QueueOrApplyStateUpdate(ctx, tx, update); err != nil {
		return nil, err
	}

	return uidWithFlags, nil
}

func (state *State) actionMoveMessagesOutOfRecoveryMailbox(
	ctx context.Context,
	tx *ent.Tx,
	messageIDs []ids.MessageIDPair,
	mboxID ids.MailboxIDPair,
) ([]db.UIDWithFlags, error) {
	ids := make([]ids.MessageIDPair, 0, len(messageIDs))
	oldInternalIDs := make([]imap.InternalMessageID, 0, len(messageIDs))

	// Import messages to remote.
	for _, id := range messageIDs {
		newID, deduped, err := state.actionImportRecoveredMessage(ctx, tx, id.InternalID, mboxID.RemoteID)
		if err != nil {
			return nil, err
		}

		if !deduped {
			if err := db.MarkMessageAsDeleted(ctx, tx, id.InternalID); err != nil {
				return nil, err
			}
		}

		ids = append(ids, newID)
		oldInternalIDs = append(oldInternalIDs, id.InternalID)
	}

	// Expunge messages
	var updates []Update
	{
		removeUpdates, err := RemoveMessagesFromMailbox(ctx, tx, state.user.GetRecoveryMailboxID().InternalID, oldInternalIDs)
		if err != nil {
			return nil, err
		}

		updates = append(updates, removeUpdates...)
	}

	// Label messages in destination.
	uidWithFlags, update, err := state.actionAddRecoveredMessagesToMailbox(ctx, tx, ids, mboxID)
	if err != nil {
		return nil, err
	}

	// Publish all updates in unison.
	updates = append(updates, update)

	if err := state.user.QueueOrApplyStateUpdate(ctx, tx, updates...); err != nil {
		return nil, err
	}

	return uidWithFlags, nil
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

	if mboxID.InternalID != state.user.GetRecoveryMailboxID().InternalID {
		if err := state.user.GetRemote().RemoveMessagesFromMailbox(ctx, remoteIDs, mboxID.RemoteID); err != nil {
			return err
		}
	}

	updates, err := RemoveMessagesFromMailbox(ctx, tx, mboxID.InternalID, internalIDs)
	if err != nil {
		return err
	}

	return state.user.QueueOrApplyStateUpdate(ctx, tx, updates...)
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

	shouldRemoveOldMessages, err := state.user.GetRemote().MoveMessagesFromMailbox(ctx, remoteIDs, mboxFromID.RemoteID, mboxToID.RemoteID)
	if err != nil {
		return nil, err
	}

	messageUIDs, updates, err := MoveMessagesFromMailbox(ctx, tx, mboxFromID.InternalID, mboxToID.InternalID, internalIDs, state, state.imapLimits, shouldRemoveOldMessages)
	if err != nil {
		return nil, err
	}

	if err := state.user.QueueOrApplyStateUpdate(ctx, tx, updates...); err != nil {
		return nil, err
	}

	return messageUIDs, nil
}

func (state *State) actionAddMessageFlags(
	ctx context.Context,
	tx *ent.Tx,
	messages []snapMsgWithSeq,
	addFlags imap.FlagSet,
) error {
	internalMessageIDs := xslices.Map(messages, func(sm snapMsgWithSeq) imap.InternalMessageID {
		return sm.ID.InternalID
	})

	if err := state.applyMessageFlagsAdded(ctx, tx, internalMessageIDs, addFlags); err != nil {
		return err
	}

	return nil
}

func (state *State) actionRemoveMessageFlags(
	ctx context.Context,
	tx *ent.Tx,
	messages []snapMsgWithSeq,
	remFlags imap.FlagSet,
) error {
	internalMessageIDs := xslices.Map(messages, func(sm snapMsgWithSeq) imap.InternalMessageID {
		return sm.ID.InternalID
	})

	if err := state.applyMessageFlagsRemoved(ctx, tx, internalMessageIDs, remFlags); err != nil {
		return err
	}

	return nil
}

func (state *State) actionSetMessageFlags(ctx context.Context, tx *ent.Tx, messages []snapMsgWithSeq, setFlags imap.FlagSet) error {
	if setFlags.ContainsUnchecked(imap.FlagRecentLowerCase) {
		return fmt.Errorf("recent flag is read-only")
	}

	internalMessageIDs := xslices.Map(messages, func(sm snapMsgWithSeq) imap.InternalMessageID {
		return sm.ID.InternalID
	})

	return state.applyMessageFlagsSet(ctx, tx, internalMessageIDs, setFlags)
}
