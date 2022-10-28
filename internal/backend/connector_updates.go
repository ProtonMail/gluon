package backend

import (
	"context"
	"fmt"
	"strings"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db"
	"github.com/ProtonMail/gluon/internal/db/ent"
	"github.com/ProtonMail/gluon/internal/ids"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/ProtonMail/gluon/rfc822"
	"github.com/bradenaw/juniper/xslices"
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

	case *imap.MessageMailboxesUpdated:
		return user.applyMessageMailboxesUpdated(ctx, update)

	case *imap.MessageFlagsUpdated:
		return user.applyMessageFlagsUpdated(ctx, update)

	case *imap.MessageIDChanged:
		return user.applyMessageIDChanged(ctx, update)

	case *imap.MessageDeleted:
		return user.applyMessageDeleted(ctx, update)

	case *imap.Noop:
		return nil

	default:
		return fmt.Errorf("bad update")
	}
}

// applyMailboxCreated applies a MailboxCreated update.
func (user *user) applyMailboxCreated(ctx context.Context, update *imap.MailboxCreated) error {
	if exists, err := db.ReadResult(ctx, user.db, func(ctx context.Context, client *ent.Client) (bool, error) {
		return db.MailboxExistsWithRemoteID(ctx, client, update.Mailbox.ID)
	}); err != nil {
		return err
	} else if exists {
		return nil
	}

	return user.db.Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		if _, err := db.CreateMailbox(
			ctx,
			tx,
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
	internalMailboxID, err := db.ReadResult(ctx, user.db, func(ctx context.Context, client *ent.Client) (imap.InternalMailboxID, error) {
		return db.GetMailboxIDFromRemoteID(ctx, client, update.MailboxID)
	})
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}

		return err
	}

	if err := user.db.Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		uidValidity, increased, err := db.DeleteMailboxWithRemoteID(ctx, tx, update.MailboxID)
		if err != nil {
			return err
		}

		if increased {
			if err := user.connector.SetUIDValidity(uidValidity); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	user.queueStateUpdate(state.NewMailboxDeletedStateUpdate(internalMailboxID))

	return nil
}

// applyMailboxUpdated applies a MailboxUpdated update.
func (user *user) applyMailboxUpdated(ctx context.Context, update *imap.MailboxUpdated) error {
	if exists, err := db.ReadResult(ctx, user.db, func(ctx context.Context, client *ent.Client) (bool, error) {
		return db.MailboxExistsWithRemoteID(ctx, client, update.MailboxID)
	}); err != nil {
		return err
	} else if !exists {
		return nil
	}

	currentName, err := db.ReadResult(ctx, user.db, func(ctx context.Context, client *ent.Client) (string, error) {
		return db.GetMailboxNameWithRemoteID(ctx, client, update.MailboxID)
	})
	if err != nil {
		return err
	}

	if currentName == strings.Join(update.MailboxName, user.delimiter) {
		return nil
	}

	return user.db.Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		return db.RenameMailboxWithRemoteID(ctx, tx, update.MailboxID, strings.Join(update.MailboxName, user.delimiter))
	})
}

// applyMailboxIDChanged applies a MailboxIDChanged update.
func (user *user) applyMailboxIDChanged(ctx context.Context, update *imap.MailboxIDChanged) error {
	return user.db.Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		if err := db.UpdateRemoteMailboxID(ctx, tx, update.InternalID, update.RemoteID); err != nil {
			return err
		}

		user.queueStateUpdate(state.NewMailboxRemoteIDUpdateStateUpdate(update.InternalID, update.RemoteID))

		return nil
	})
}

// applyMessagesCreated applies a MessagesCreated update.
func (user *user) applyMessagesCreated(ctx context.Context, update *imap.MessagesCreated) error {
	// collect all unique messages to create
	messagesToCreate := make([]*db.CreateMessageReq, 0, len(update.Messages))
	messagesToCreateFilter := make(map[imap.MessageID]imap.InternalMessageID, len(update.Messages)/2)
	messageForMBox := make(map[imap.InternalMailboxID][]imap.InternalMessageID)
	mboxInternalIDMap := make(map[imap.MailboxID]imap.InternalMailboxID)

	if err := user.db.Read(ctx, func(ctx context.Context, client *ent.Client) error {
		for _, message := range update.Messages {
			internalID, ok := messagesToCreateFilter[message.Message.ID]
			if !ok {
				_, err := db.GetMessageIDFromRemoteID(ctx, client, message.Message.ID)
				if ent.IsNotFound(err) {
					internalID = user.nextMessageID()
				} else {
					return err
				}

				literal, err := rfc822.SetHeaderValue(message.Literal, ids.InternalIDKey, internalID.String())
				if err != nil {
					return fmt.Errorf("failed to set internal ID: %w", err)
				}

				request := &db.CreateMessageReq{
					Message:    message.Message,
					Literal:    literal,
					Body:       message.ParsedMessage.Body,
					Structure:  message.ParsedMessage.Structure,
					Envelope:   message.ParsedMessage.Envelope,
					InternalID: internalID,
				}

				messagesToCreate = append(messagesToCreate, request)
				messagesToCreateFilter[message.Message.ID] = internalID
			}

			for _, mboxID := range message.MailboxIDs {
				v, ok := mboxInternalIDMap[mboxID]
				if !ok {
					internalMBoxID, err := db.GetMailboxIDWithRemoteID(ctx, client, mboxID)
					if err != nil {
						return err
					}

					v = internalMBoxID
					mboxInternalIDMap[mboxID] = v
				}

				messageList, ok := messageForMBox[v]
				if !ok {
					messageList = []imap.InternalMessageID{}
					messageForMBox[v] = messageList
				}

				if !slices.Contains(messageList, internalID) {
					messageList = append(messageList, internalID)
					messageForMBox[v] = messageList
				}
			}
		}
		return nil
	}); err != nil {
		return err
	}

	if len(messagesToCreate) == 0 && len(messageForMBox) == 0 {
		return nil
	}

	// We sadly have to split this up into two separate transactions where we create the messages and one where we
	// assign them to the mailbox. There's an upper limit to the number of items badger can track in one transaction.
	// This way we can keep the database consistent.
	for _, chunk := range xslices.Chunk(messagesToCreate, db.ChunkLimit) {
		if err := user.db.Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
			// Create messages in the store
			for _, msg := range chunk {
				if err := user.store.Set(msg.InternalID, msg.Literal); err != nil {
					return fmt.Errorf("failed to store message literal: %w", err)
				}
			}

			// Create message in the database
			if _, err := db.CreateMessages(ctx, tx, chunk...); err != nil {
				return err
			}

			return nil
		}); err != nil {
			return err
		}
	}

	return user.db.Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		// Assign all the messages to the mailbox
		for mboxID, msgList := range messageForMBox {
			if _, err := user.applyMessagesAddedToMailbox(ctx, tx, mboxID, msgList); err != nil {
				return err
			}
		}

		return nil
	})
}

// applyMessageMailboxesUpdated applies a MessageMailboxesUpdated update.
func (user *user) applyMessageMailboxesUpdated(ctx context.Context, update *imap.MessageMailboxesUpdated) error {
	if exists, err := db.ReadResult(ctx, user.db, func(ctx context.Context, client *ent.Client) (bool, error) {
		return db.MessageExistsWithRemoteID(ctx, client, update.MessageID)
	}); err != nil {
		return err
	} else if !exists {
		return state.ErrNoSuchMessage
	}

	type Result struct {
		InternalMsgID   imap.InternalMessageID
		InternalMBoxIDs []imap.InternalMailboxID
	}

	result, err := db.ReadResult(ctx, user.db, func(ctx context.Context, client *ent.Client) (Result, error) {
		internalMsgID, err := db.GetMessageIDFromRemoteID(ctx, client, update.MessageID)
		if err != nil {
			return Result{}, err
		}

		internalMBoxIDs, err := db.TranslateRemoteMailboxIDs(ctx, client, update.MailboxIDs)
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
	if exists, err := db.ReadResult(ctx, user.db, func(ctx context.Context, client *ent.Client) (bool, error) {
		return db.MessageExistsWithRemoteID(ctx, client, update.MessageID)
	}); err != nil {
		return err
	} else if !exists {
		return state.ErrNoSuchMessage
	}

	internalMsgID, err := db.ReadResult(ctx, user.db, func(ctx context.Context, client *ent.Client) (imap.InternalMessageID, error) {
		return db.GetMessageIDFromRemoteID(ctx, client, update.MessageID)
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
		return db.UpdateRemoteMessageID(ctx, tx, update.InternalID, update.RemoteID)
	}); err != nil {
		return err
	}

	if err := user.forState(func(state *state.State) error {
		return state.UpdateMessageRemoteID(update.InternalID, update.RemoteID)
	}); err != nil {
		return err
	}

	return nil
}

func (user *user) setMessageMailboxes(ctx context.Context, tx *ent.Tx, messageID imap.InternalMessageID, mboxIDs []imap.InternalMailboxID) error {
	curMailboxIDs, err := db.GetMessageMailboxIDs(ctx, tx.Client(), messageID)
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
func (user *user) applyMessagesAddedToMailbox(ctx context.Context, tx *ent.Tx, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) ([]db.UIDWithFlags, error) {
	messageUIDs, update, err := state.AddMessagesToMailbox(ctx, tx, mboxID, messageIDs, nil)
	if err != nil {
		return nil, err
	}

	user.queueStateUpdate(update)

	return messageUIDs, nil
}

// applyMessagesRemovedFromMailbox removes the messages from the given mailbox.
func (user *user) applyMessagesRemovedFromMailbox(ctx context.Context, tx *ent.Tx, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) error {
	updates, err := state.RemoveMessagesFromMailbox(ctx, tx, mboxID, messageIDs)
	if err != nil {
		return err
	}

	for _, update := range updates {
		user.queueStateUpdate(update)
	}

	return nil
}

func (user *user) setMessageFlags(ctx context.Context, tx *ent.Tx, messageID imap.InternalMessageID, seen, flagged bool) error {
	curFlags, err := db.GetMessageFlags(ctx, tx.Client(), []imap.InternalMessageID{messageID})
	if err != nil {
		return err
	}

	flagSet := curFlags[0].FlagSet

	if seen && !flagSet.ContainsUnchecked(imap.FlagSeenLowerCase) {
		if err := user.addMessageFlags(ctx, tx, messageID, imap.FlagSeen); err != nil {
			return err
		}
	} else if !seen && flagSet.ContainsUnchecked(imap.FlagSeenLowerCase) {
		if err := user.removeMessageFlags(ctx, tx, messageID, imap.FlagSeen); err != nil {
			return err
		}
	}

	if flagged && !flagSet.ContainsUnchecked(imap.FlagFlaggedLowerCase) {
		if err := user.addMessageFlags(ctx, tx, messageID, imap.FlagFlagged); err != nil {
			return err
		}
	} else if !flagged && flagSet.ContainsUnchecked(imap.FlagFlaggedLowerCase) {
		if err := user.removeMessageFlags(ctx, tx, messageID, imap.FlagFlagged); err != nil {
			return err
		}
	}

	return nil
}

func (user *user) addMessageFlags(ctx context.Context, tx *ent.Tx, messageID imap.InternalMessageID, flag string) error {
	if err := db.AddMessageFlag(ctx, tx, []imap.InternalMessageID{messageID}, flag); err != nil {
		return err
	}

	user.queueStateUpdate(state.NewRemoteAddMessageFlagsStateUpdate(messageID, flag))

	return nil
}

func (user *user) removeMessageFlags(ctx context.Context, tx *ent.Tx, messageID imap.InternalMessageID, flag string) error {
	if err := db.RemoveMessageFlag(ctx, tx, []imap.InternalMessageID{messageID}, flag); err != nil {
		return err
	}

	user.queueStateUpdate(state.NewRemoteRemoveMessageFlagsStateUpdate(messageID, flag))

	return nil
}

func (user *user) applyMessageDeleted(ctx context.Context, update *imap.MessageDeleted) error {
	var stateUpdates []state.Update

	if err := user.db.Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		if err := db.MarkMessageAsDeletedWithRemoteID(ctx, tx, update.MessageID); err != nil {
			return err
		}

		internalMessageID, err := db.GetMessageIDFromRemoteID(ctx, tx.Client(), update.MessageID)
		if err != nil {
			return err
		}

		mailboxes, err := db.GetMessageMailboxIDs(ctx, tx.Client(), internalMessageID)
		if err != nil {
			return err
		}

		messageIDs := []imap.InternalMessageID{internalMessageID}

		for _, mailbox := range mailboxes {
			updates, err := state.RemoveMessagesFromMailbox(ctx, tx, mailbox, messageIDs)
			if err != nil {
				return err
			}

			stateUpdates = append(stateUpdates, updates...)
		}

		return nil
	}); err != nil {
		return err
	}

	for _, update := range stateUpdates {
		user.queueStateUpdate(update)
	}

	return nil
}
