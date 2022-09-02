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
	"github.com/ProtonMail/gluon/store"
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
	if exists, err := db.ReadResult(ctx, user.db, func(ctx context.Context, client *ent.Client) (bool, error) {
		return db.MailboxExistsWithRemoteID(ctx, client, update.Mailbox.ID)
	}); err != nil {
		return err
	} else if exists {
		return nil
	}

	internalMailboxID := imap.InternalMailboxID(uuid.NewString())

	return user.db.Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		if _, err := db.CreateMailbox(
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
	if exists, err := db.ReadResult(ctx, user.db, func(ctx context.Context, client *ent.Client) (bool, error) {
		return db.MailboxExistsWithRemoteID(ctx, client, update.MailboxID)
	}); err != nil {
		return err
	} else if !exists {
		return nil
	}

	return user.db.Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		return db.DeleteMailboxWithRemoteID(ctx, tx, update.MailboxID)
	})
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
	var updates []*imap.MessageCreated

	if err := user.db.Read(ctx, func(ctx context.Context, client *ent.Client) error {
		for _, update := range update.Messages {
			if exists, err := db.MessageExistsWithRemoteID(ctx, client, update.Message.ID); err != nil {
				return err
			} else if !exists {
				updates = append(updates, update)
			}
		}

		return nil
	}); err != nil {
		return err
	}

	var reqs []*db.CreateMessageReq

	remoteToLocalMessageID := make(map[imap.MessageID]imap.InternalMessageID)

	return db.WriteAndStore(ctx, user.db, user.store, func(ctx context.Context, tx *ent.Tx, storeTx store.Transaction) error {
		for _, update := range updates {
			internalID := uuid.NewString()

			literal, err := rfc822.SetHeaderValue(update.Literal, ids.InternalIDKey, internalID)
			if err != nil {
				return fmt.Errorf("failed to set internal ID: %w", err)
			}

			if err := storeTx.Set(imap.InternalMessageID(internalID), literal); err != nil {
				return fmt.Errorf("failed to store message literal: %w", err)
			}

			reqs = append(reqs, &db.CreateMessageReq{
				Message:    update.Message,
				Literal:    literal,
				Body:       update.Body,
				Structure:  update.Structure,
				Envelope:   update.Envelope,
				InternalID: imap.InternalMessageID(internalID),
			})

			remoteToLocalMessageID[update.Message.ID] = imap.InternalMessageID(internalID)
		}

		if _, err := db.CreateMessages(ctx, tx, reqs...); err != nil {
			return fmt.Errorf("failed to create message: %w", err)
		}

		messageIDs := make(map[imap.LabelID][]ids.MessageIDPair)

		for _, update := range updates {
			for _, mailboxID := range update.MailboxIDs {
				localID := remoteToLocalMessageID[update.Message.ID]
				idPair := ids.MessageIDPair{
					InternalID: localID,
					RemoteID:   update.Message.ID,
				}

				if !slices.Contains(messageIDs[mailboxID], idPair) {
					messageIDs[mailboxID] = append(messageIDs[mailboxID], idPair)
				}
			}
		}

		for mailboxID, messageIDs := range messageIDs {
			internalMailboxID, err := db.GetMailboxIDWithRemoteID(ctx, tx.Client(), mailboxID)
			if err != nil {
				return err
			}

			internalIDs := xslices.Map(messageIDs, func(id ids.MessageIDPair) imap.InternalMessageID {
				return id.InternalID
			})

			messageUIDs, err := db.AddMessagesToMailbox(ctx, tx, internalIDs, internalMailboxID)
			if err != nil {
				return err
			}

			responders := xslices.Map(messageIDs, func(messageID ids.MessageIDPair) state.Responder {
				return state.NewExists(messageID.InternalID, messageUIDs[messageID.InternalID])
			})

			user.queueStateUpdate(state.NewMailboxIDResponderStateUpdate(internalMailboxID, responders...))
		}

		return nil
	})
}

// applyMessageLabelsUpdated applies a MessageLabelsUpdated update.
func (user *user) applyMessageLabelsUpdated(ctx context.Context, update *imap.MessageLabelsUpdated) error {
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
func (user *user) applyMessagesAddedToMailbox(ctx context.Context, tx *ent.Tx, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) (map[imap.InternalMessageID]int, error) {
	messageUIDs, update, err := state.AddMessagesToMailbox(ctx, tx, mboxID, messageIDs)
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
	return user.db.Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		if err := db.MarkMessageAsDeletedWithRemoteID(ctx, tx, update.MessageID); err != nil {
			return err
		}

		internalMessageID, err := db.GetMessageIDFromRemoteID(ctx, tx.Client(), update.MessageID)
		if err != nil {
			return err
		}

		user.queueStateUpdate(state.NewRemoteMessageDeletedStateUpdate(internalMessageID, update.MessageID))

		return nil
	})
}
