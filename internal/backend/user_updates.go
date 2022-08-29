package backend

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/backend/ent"
	"github.com/ProtonMail/gluon/rfc822"
	"github.com/bradenaw/juniper/xslices"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

// apply an incoming update originating from the connector.
func (user *user) apply(ctx context.Context, update imap.Update) error {
	defer update.Done()

	logrus.WithField("update", update).Debug("Applying update")

	switch update := update.(type) {
	case *imap.MailboxCreated:
		return user.applyMailboxCreated(ctx, update)

	case *imap.MailboxDeleted:
		return user.applyMailboxDeleted(ctx, update)

	case *imap.MailboxUpdated:
		return user.applyMailboxUpdated(ctx, update)

	case *imap.MailboxIDChanged:
		return user.applyMailboxIDChanged(ctx, update)

	case *imap.MessagesCreated:
		return user.applyMessagesCreated(ctx, update)

	case *imap.MessageLabelsUpdated:
		return user.applyMessageLabelsUpdated(ctx, update)

	case *imap.MessageFlagsUpdated:
		return user.applyMessageFlagsUpdated(ctx, update)

	case *imap.MessageIDChanged:
		return user.applyMessageIDChanged(ctx, update)

	case *imap.MessageDeleted:
		return user.applyMessageDeleted(ctx, update)

	default:
		return fmt.Errorf("bad update")
	}
}

// applyMailboxCreated applies a MailboxCreated update.
func (user *user) applyMailboxCreated(ctx context.Context, update *imap.MailboxCreated) error {
	if exists, err := DBReadResult(ctx, user.db, func(ctx context.Context, client *ent.Client) (bool, error) {
		return DBMailboxExistsWithRemoteID(ctx, client, update.Mailbox.ID)
	}); err != nil {
		return err
	} else if exists {
		return nil
	}

	internalMailboxID := imap.InternalMailboxID(uuid.NewString())

	return user.db.Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		if _, err := DBCreateMailbox(
			ctx,
			tx,
			internalMailboxID,
			update.Mailbox.ID,
			strings.Join(update.Mailbox.Name, user.delimiter),
			update.Mailbox.Flags,
			update.Mailbox.PermanentFlags,
			update.Mailbox.Attributes,
		); err != nil {
			return err
		}

		return nil
	})
}

// applyMailboxDeleted applies a MailboxDeleted update.
func (user *user) applyMailboxDeleted(ctx context.Context, update *imap.MailboxDeleted) error {
	if exists, err := DBReadResult(ctx, user.db, func(ctx context.Context, client *ent.Client) (bool, error) {
		return DBMailboxExistsWithRemoteID(ctx, client, update.MailboxID)
	}); err != nil {
		return err
	} else if !exists {
		return nil
	}

	return user.db.Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		return DBDeleteMailboxWithRemoteID(ctx, tx, update.MailboxID)
	})
}

// applyMailboxUpdated applies a MailboxUpdated update.
func (user *user) applyMailboxUpdated(ctx context.Context, update *imap.MailboxUpdated) error {
	if exists, err := DBReadResult(ctx, user.db, func(ctx context.Context, client *ent.Client) (bool, error) {
		return DBMailboxExistsWithRemoteID(ctx, client, update.MailboxID)
	}); err != nil {
		return err
	} else if !exists {
		return nil
	}

	currentName, err := DBReadResult(ctx, user.db, func(ctx context.Context, client *ent.Client) (string, error) {
		return DBGetMailboxNameWithRemoteID(ctx, client, update.MailboxID)
	})
	if err != nil {
		return err
	}

	if currentName == strings.Join(update.MailboxName, user.delimiter) {
		return nil
	}

	return user.db.Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		return DBRenameMailboxWithRemoteID(ctx, tx, update.MailboxID, strings.Join(update.MailboxName, user.delimiter))
	})
}

// applyMailboxIDChanged applies a MailboxIDChanged update.
func (user *user) applyMailboxIDChanged(ctx context.Context, update *imap.MailboxIDChanged) error {
	return user.db.Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		if err := DBUpdateRemoteMailboxID(ctx, tx, update.InternalID, update.RemoteID); err != nil {
			return err
		}

		if err := user.queueOrApplyStateUpdate(ctx, tx, newMailboxRemoteIDUpdateStateUpdate(update.InternalID, update.RemoteID)); err != nil {
			return err
		}

		return nil
	})
}

// applyMessagesCreated applies a MessagesCreated update.
func (user *user) applyMessagesCreated(ctx context.Context, update *imap.MessagesCreated) error {
	var updates []*imap.MessageCreated

	if err := user.db.Read(ctx, func(ctx context.Context, client *ent.Client) error {
		for _, update := range update.Messages {
			if exists, err := DBMessageExistsWithRemoteID(ctx, client, update.Message.ID); err != nil {
				return err
			} else if !exists {
				updates = append(updates, update)
			}
		}

		return nil
	}); err != nil {
		return err
	}

	var reqs []*DBCreateMessageReq

	remoteToLocalMessageID := make(map[imap.MessageID]imap.InternalMessageID)

	for _, update := range updates {
		internalID := uuid.NewString()

		literal, err := rfc822.SetHeaderValue(update.Literal, InternalIDKey, internalID)
		if err != nil {
			return fmt.Errorf("failed to set internal ID: %w", err)
		}

		if err := user.store.Set(imap.InternalMessageID(internalID), literal); err != nil {
			return fmt.Errorf("failed to store message literal: %w", err)
		}

		reqs = append(reqs, &DBCreateMessageReq{
			message:    update.Message,
			literal:    literal,
			body:       update.Body,
			structure:  update.Structure,
			envelope:   update.Envelope,
			internalID: imap.InternalMessageID(internalID),
		})

		remoteToLocalMessageID[update.Message.ID] = imap.InternalMessageID(internalID)
	}

	return user.db.Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		if _, err := DBCreateMessages(ctx, tx, reqs...); err != nil {
			return fmt.Errorf("failed to create message: %w", err)
		}

		messageIDs := make(map[imap.LabelID][]MessageIDPair)

		for _, update := range updates {
			for _, mailboxID := range update.MailboxIDs {
				localID := remoteToLocalMessageID[update.Message.ID]
				idPair := MessageIDPair{
					InternalID: localID,
					RemoteID:   update.Message.ID,
				}

				if !slices.Contains(messageIDs[mailboxID], idPair) {
					messageIDs[mailboxID] = append(messageIDs[mailboxID], idPair)
				}
			}
		}

		for mailboxID, messageIDs := range messageIDs {
			internalMailboxID, err := DBGetMailboxIDWithRemoteID(ctx, tx.Client(), mailboxID)
			if err != nil {
				return err
			}

			internalIDs := xslices.Map(messageIDs, func(id MessageIDPair) imap.InternalMessageID {
				return id.InternalID
			})

			messageUIDs, err := DBAddMessagesToMailbox(ctx, tx, internalIDs, internalMailboxID)
			if err != nil {
				return err
			}

			responders := xslices.Map(messageIDs, func(messageID MessageIDPair) responder {
				return newExists(messageID.InternalID, messageUIDs[messageID.InternalID])
			})

			if err := user.queueOrApplyStateUpdate(ctx, tx, newMailboxIDResponderStateUpdate(internalMailboxID, responders...)); err != nil {
				return err
			}
		}

		return nil
	})
}

// applyMessageLabelsUpdated applies a MessageLabelsUpdated update.
func (user *user) applyMessageLabelsUpdated(ctx context.Context, update *imap.MessageLabelsUpdated) error {
	if exists, err := DBReadResult(ctx, user.db, func(ctx context.Context, client *ent.Client) (bool, error) {
		return DBMessageExistsWithRemoteID(ctx, client, update.MessageID)
	}); err != nil {
		return err
	} else if !exists {
		return ErrNoSuchMessage
	}

	type Result struct {
		InternalMsgID   imap.InternalMessageID
		InternalMBoxIDs []imap.InternalMailboxID
	}

	result, err := DBReadResult(ctx, user.db, func(ctx context.Context, client *ent.Client) (Result, error) {
		internalMsgID, err := DBGetMessageIDFromRemoteID(ctx, client, update.MessageID)
		if err != nil {
			return Result{}, err
		}

		internalMBoxIDs, err := DBTranslateRemoteMailboxIDs(ctx, client, update.MailboxIDs)
		if err != nil {
			return Result{}, err
		}

		return Result{
			InternalMsgID:   internalMsgID,
			InternalMBoxIDs: internalMBoxIDs,
		}, nil
	})
	if err != nil {
		return err
	}

	return user.db.Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		if err := user.setMessageMailboxes(ctx, tx, result.InternalMsgID, result.InternalMBoxIDs); err != nil {
			return err
		}

		if err := user.setMessageFlags(ctx, tx, result.InternalMsgID, update.Seen, update.Flagged); err != nil {
			return err
		}

		return nil
	})
}

// applyMessageFlagsUpdated applies a MessageFlagsUpdated update.
func (user *user) applyMessageFlagsUpdated(ctx context.Context, update *imap.MessageFlagsUpdated) error {
	if exists, err := DBReadResult(ctx, user.db, func(ctx context.Context, client *ent.Client) (bool, error) {
		return DBMessageExistsWithRemoteID(ctx, client, update.MessageID)
	}); err != nil {
		return err
	} else if !exists {
		return ErrNoSuchMessage
	}

	internalMsgID, err := DBReadResult(ctx, user.db, func(ctx context.Context, client *ent.Client) (imap.InternalMessageID, error) {
		return DBGetMessageIDFromRemoteID(ctx, client, update.MessageID)
	})
	if err != nil {
		return err
	}

	return user.db.Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		if err := user.setMessageFlags(ctx, tx, internalMsgID, update.Seen, update.Flagged); err != nil {
			return err
		}

		return nil
	})
}

// applyMessageIDChanged applies a MessageIDChanged update.
func (user *user) applyMessageIDChanged(ctx context.Context, update *imap.MessageIDChanged) error {
	if err := user.db.Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		return DBUpdateRemoteMessageID(ctx, tx, update.InternalID, update.RemoteID)
	}); err != nil {
		return err
	}

	if err := user.forState(func(state *State) error {
		return state.updateMessageRemoteID(update.InternalID, update.RemoteID)
	}); err != nil {
		return err
	}

	return nil
}

func (user *user) setMessageMailboxes(ctx context.Context, tx *ent.Tx, messageID imap.InternalMessageID, mboxIDs []imap.InternalMailboxID) error {
	curMailboxIDs, err := DBGetMessageMailboxIDs(ctx, tx.Client(), messageID)
	if err != nil {
		return err
	}

	for _, mboxID := range xslices.Filter(mboxIDs, func(mboxID imap.InternalMailboxID) bool { return !slices.Contains(curMailboxIDs, mboxID) }) {
		if _, err := user.applyMessagesAddedToMailbox(ctx, tx, mboxID, []imap.InternalMessageID{messageID}); err != nil {
			return err
		}
	}

	for _, mboxID := range xslices.Filter(curMailboxIDs, func(mboxID imap.InternalMailboxID) bool { return !slices.Contains(mboxIDs, mboxID) }) {
		if err := user.applyMessagesRemovedFromMailbox(ctx, tx, mboxID, []imap.InternalMessageID{messageID}); err != nil {
			return err
		}
	}

	return nil
}

// applyMessagesAddedToMailbox adds the messages to the given mailbox.
func (user *user) applyMessagesAddedToMailbox(ctx context.Context, tx *ent.Tx, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) (map[imap.InternalMessageID]int, error) {
	if _, err := DBAddMessagesToMailbox(ctx, tx, messageIDs, mboxID); err != nil {
		return nil, err
	}

	messageUIDs, err := DBGetMessageUIDs(ctx, tx.Client(), mboxID, messageIDs)
	if err != nil {
		return nil, err
	}

	responders := xslices.Map(messageIDs, func(id imap.InternalMessageID) responder {
		return newExists(id, messageUIDs[id])
	})

	if err := user.queueOrApplyStateUpdate(ctx, tx, newMailboxIDResponderStateUpdate(mboxID, responders...)); err != nil {
		return nil, err
	}

	return messageUIDs, nil
}

// applyMessagesRemovedFromMailbox removes the messages from the given mailbox.
func (user *user) applyMessagesRemovedFromMailbox(ctx context.Context, tx *ent.Tx, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) error {
	if len(messageIDs) > 0 {
		if err := DBRemoveMessagesFromMailbox(ctx, tx, messageIDs, mboxID); err != nil {
			return err
		}
	}

	for _, messageID := range messageIDs {
		if err := user.queueOrApplyStateUpdate(ctx, tx, newMessageIDAndMailboxIDResponderStateUpdate(messageID, mboxID, newExpunge(messageID, isClose(ctx)))); err != nil {
			return err
		}
	}

	return nil
}

func (user *user) applyMessagesMovedFromMailbox(
	ctx context.Context,
	tx *ent.Tx,
	mboxFromID, mboxToID imap.InternalMailboxID,
	messageIDs []imap.InternalMessageID,
) (map[imap.InternalMessageID]int, error) {
	if mboxFromID != mboxToID {
		if err := DBRemoveMessagesFromMailbox(ctx, tx, messageIDs, mboxFromID); err != nil {
			return nil, err
		}
	}

	if _, err := DBAddMessagesToMailbox(ctx, tx, messageIDs, mboxToID); err != nil {
		return nil, err
	}

	messageUIDs, err := DBGetMessageUIDs(ctx, tx.Client(), mboxToID, messageIDs)
	if err != nil {
		return nil, err
	}

	{
		responders := xslices.Map(messageIDs, func(id imap.InternalMessageID) responder {
			return newExists(id, messageUIDs[id])
		})
		if err := user.queueOrApplyStateUpdate(ctx, tx, newMailboxIDResponderStateUpdate(mboxToID, responders...)); err != nil {
			return nil, err
		}
	}

	for _, messageID := range messageIDs {
		if err := user.queueOrApplyStateUpdate(ctx, tx, newMessageIDAndMailboxIDResponderStateUpdate(messageID, mboxFromID, newExpunge(messageID, isClose(ctx)))); err != nil {
			return nil, err
		}
	}

	return messageUIDs, nil
}

func (user *user) setMessageFlags(ctx context.Context, tx *ent.Tx, messageID imap.InternalMessageID, seen, flagged bool) error {
	curFlags, err := DBGetMessageFlags(ctx, tx.Client(), []imap.InternalMessageID{messageID})
	if err != nil {
		return err
	}

	if seen && !curFlags[messageID].Contains(imap.FlagSeen) {
		if err := user.addMessageFlags(ctx, tx, messageID, imap.FlagSeen); err != nil {
			return err
		}
	} else if !seen && curFlags[messageID].Contains(imap.FlagSeen) {
		if err := user.removeMessageFlags(ctx, tx, messageID, imap.FlagSeen); err != nil {
			return err
		}
	}

	if flagged && !curFlags[messageID].Contains(imap.FlagFlagged) {
		if err := user.addMessageFlags(ctx, tx, messageID, imap.FlagFlagged); err != nil {
			return err
		}
	} else if !flagged && curFlags[messageID].Contains(imap.FlagFlagged) {
		if err := user.removeMessageFlags(ctx, tx, messageID, imap.FlagFlagged); err != nil {
			return err
		}
	}

	return nil
}

type remoteAddMessageFlagsStateUpdate struct {
	messageIDStateFilter
	flag string
}

func NewRemoteAddMessageFlagsStateUpdate(messageID imap.InternalMessageID, flag string) stateUpdate {
	return &remoteAddMessageFlagsStateUpdate{
		messageIDStateFilter: messageIDStateFilter{messageID: messageID},
		flag:                 flag,
	}
}

func (u *remoteAddMessageFlagsStateUpdate) apply(ctx context.Context, tx *ent.Tx, s *State) error {
	snapFlags, err := s.snap.getMessageFlags(u.messageID)
	if err != nil {
		return err
	}

	return s.pushResponder(ctx, tx, newFetch(u.messageID, snapFlags.Add(u.flag), isUID(ctx), isSilent(ctx)))
}

func (user *user) addMessageFlags(ctx context.Context, tx *ent.Tx, messageID imap.InternalMessageID, flag string) error {
	if err := DBAddMessageFlag(ctx, tx, []imap.InternalMessageID{messageID}, flag); err != nil {
		return err
	}

	return user.queueOrApplyStateUpdate(ctx, tx, NewRemoteAddMessageFlagsStateUpdate(messageID, flag))
}

type remoteRemoveMessageFlagsStateUpdate struct {
	messageIDStateFilter
	flag string
}

func NewRemoteRemoveMessageFlagsStateUpdate(messageID imap.InternalMessageID, flag string) stateUpdate {
	return &remoteRemoveMessageFlagsStateUpdate{
		messageIDStateFilter: messageIDStateFilter{messageID: messageID},
		flag:                 flag,
	}
}

func (u *remoteRemoveMessageFlagsStateUpdate) apply(ctx context.Context, tx *ent.Tx, s *State) error {
	snapFlags, err := s.snap.getMessageFlags(u.messageID)
	if err != nil {
		if errors.Is(err, ErrNoSuchMessage) {
			return nil
		}

		return err
	}

	return s.pushResponder(ctx, tx, newFetch(u.messageID, snapFlags.Remove(u.flag), isUID(ctx), isSilent(ctx)))
}

func (user *user) removeMessageFlags(ctx context.Context, tx *ent.Tx, messageID imap.InternalMessageID, flag string) error {
	if err := DBRemoveMessageFlag(ctx, tx, []imap.InternalMessageID{messageID}, flag); err != nil {
		return err
	}

	return user.queueOrApplyStateUpdate(ctx, tx, NewRemoteRemoveMessageFlagsStateUpdate(messageID, flag))
}

type remoteMessageDeletedStateUpdate struct {
	messageIDStateFilter
	remoteID imap.MessageID
}

func newRemoteMessageDeletedStateUpdate(messageID imap.InternalMessageID, remoteID imap.MessageID) stateUpdate {
	return &remoteMessageDeletedStateUpdate{
		messageIDStateFilter: messageIDStateFilter{messageID: messageID},
		remoteID:             remoteID,
	}
}

func (u *remoteMessageDeletedStateUpdate) apply(ctx context.Context, tx *ent.Tx, s *State) error {
	return s.actionRemoveMessagesFromMailbox(ctx, tx, []MessageIDPair{{
		InternalID: u.messageID,
		RemoteID:   u.remoteID,
	}}, s.snap.mboxID)
}

func (user *user) applyMessageDeleted(ctx context.Context, update *imap.MessageDeleted) error {
	return user.db.Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		if err := DBMarkMessageAsDeletedWithRemoteID(ctx, tx, update.MessageID); err != nil {
			return err
		}

		internalMessageID, err := DBGetMessageIDFromRemoteID(ctx, tx.Client(), update.MessageID)
		if err != nil {
			return err
		}

		return user.queueOrApplyStateUpdate(ctx, tx, newRemoteMessageDeletedStateUpdate(internalMessageID, update.MessageID))
	})
}
