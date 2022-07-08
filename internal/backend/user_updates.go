package backend

import (
	"context"
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

// apply applies an update, coming from the given source state. if source is nil, the update is external.
func (user *user) apply(ctx context.Context, tx *ent.Tx, update imap.Update) error {
	logrus.WithField("update", update).Debug("Applying update")

	switch update := update.(type) {
	case *imap.MailboxCreated:
		return user.applyMailboxCreated(ctx, tx, update)

	case *imap.MailboxDeleted:
		return user.applyMailboxDeleted(ctx, tx, update)

	case *imap.MailboxUpdated:
		return user.applyMailboxUpdated(ctx, tx, update)

	case *imap.MailboxIDChanged:
		return user.applyMailboxIDChanged(ctx, tx, update)

	case *imap.MessagesCreated:
		return user.applyMessagesCreated(ctx, tx, update)

	case *imap.MessageUpdated:
		return user.applyMessageUpdated(ctx, tx, update)

	case *imap.MessageIDChanged:
		return user.applyMessageIDChanged(ctx, tx, update)

	case *imap.MessageDeleted:
		return user.applyMessageDeleted(ctx, tx, update)

	default:
		panic("bad update")
	}
}

// applyMailboxCreated applies a MailboxCreated update.
func (user *user) applyMailboxCreated(ctx context.Context, tx *ent.Tx, update *imap.MailboxCreated) error {
	if exists, err := DBMailboxExistsWithID(ctx, tx.Client(), update.Mailbox.ID); err != nil {
		return err
	} else if exists {
		return nil
	}

	if _, err := DBCreateMailbox(
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
}

// applyMailboxDeleted applies a MailboxDeleted update.
func (user *user) applyMailboxDeleted(ctx context.Context, tx *ent.Tx, update *imap.MailboxDeleted) error {
	if exists, err := DBMailboxExistsWithID(ctx, tx.Client(), update.MailboxID); err != nil {
		return err
	} else if !exists {
		return nil
	}

	return DBDeleteMailbox(ctx, tx, update.MailboxID)
}

// applyMailboxUpdated applies a MailboxUpdated update.
func (user *user) applyMailboxUpdated(ctx context.Context, tx *ent.Tx, update *imap.MailboxUpdated) error {
	client := tx.Client()
	if exists, err := DBMailboxExistsWithID(ctx, client, update.MailboxID); err != nil {
		return err
	} else if !exists {
		return nil
	}

	currentName, err := DBGetMailboxName(ctx, client, update.MailboxID)
	if err != nil {
		return err
	}

	if currentName == strings.Join(update.MailboxName, user.delimiter) {
		return nil
	}

	return DBRenameMailbox(ctx, tx, update.MailboxID, strings.Join(update.MailboxName, user.delimiter))
}

// applyMailboxIDChanged applies a MailboxIDChanged update.
func (user *user) applyMailboxIDChanged(ctx context.Context, tx *ent.Tx, update *imap.MailboxIDChanged) error {
	if err := user.forStateInMailbox(update.OldID, func(state *State) error {
		return state.updateMailboxID(update.OldID, update.NewID)
	}); err != nil {
		return err
	}

	if err := user.remote.FinishMailboxIDUpdate(update.OldID); err != nil {
		logrus.WithError(err).Errorf("Call to FinishMailboxIDUpdate() failed")
	}

	return DBUpdateMailboxID(ctx, tx, update.OldID, update.NewID)
}

// applyMessagesCreated applies a MessagesCreated update.
func (user *user) applyMessagesCreated(ctx context.Context, tx *ent.Tx, update *imap.MessagesCreated) error {
	var updates []*imap.MessageCreated

	for _, update := range update.Messages {
		if exists, err := DBMessageExists(ctx, tx.Client(), update.Message.ID); err != nil {
			return err
		} else if !exists {
			updates = append(updates, update)
		}
	}

	var reqs []*DBCreateMessageReq

	for _, update := range updates {
		internalID := uuid.NewString()

		literal, err := rfc822.SetHeaderValue(update.Literal, InternalIDKey, internalID)
		if err != nil {
			return fmt.Errorf("failed to set internal ID: %w", err)
		}

		if err := user.store.Set(update.Message.ID, literal); err != nil {
			return fmt.Errorf("failed to store message literal: %w", err)
		}

		reqs = append(reqs, &DBCreateMessageReq{
			message:    update.Message,
			literal:    literal,
			body:       update.Body,
			structure:  update.Structure,
			envelope:   update.Envelope,
			internalID: internalID,
		})
	}

	if _, err := DBCreateMessages(ctx, tx, reqs...); err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	messageIDs := make(map[string][]string)

	for _, update := range updates {
		for _, mailboxID := range update.MailboxIDs {
			if !slices.Contains(messageIDs[mailboxID], update.Message.ID) {
				messageIDs[mailboxID] = append(messageIDs[mailboxID], update.Message.ID)
			}
		}
	}

	for mailboxID, messageIDs := range messageIDs {
		messageUIDs, err := DBAddMessagesToMailbox(ctx, tx, messageIDs, mailboxID)
		if err != nil {
			return err
		}

		if err := user.forStateInMailbox(mailboxID, func(state *State) error {
			return state.pushResponder(ctx, tx, xslices.Map(messageIDs, func(messageID string) responder {
				return newExists(messageID, messageUIDs[messageID])
			})...)
		}); err != nil {
			return err
		}
	}

	return nil
}

// applyMessageUpdated applies a MessageUpdated update.
func (user *user) applyMessageUpdated(ctx context.Context, tx *ent.Tx, update *imap.MessageUpdated) error {
	if exists, err := DBMessageExists(ctx, tx.Client(), update.MessageID); err != nil {
		return err
	} else if !exists {
		return ErrNoSuchMessage
	}

	if err := user.setMessageMailboxes(ctx, tx, update.MessageID, update.MailboxIDs); err != nil {
		return err
	}

	if err := user.setMessageFlags(ctx, tx, update.MessageID, update.Seen, update.Flagged); err != nil {
		return err
	}

	return nil
}

// applyMessageIDChanged applies a MessageIDChanged update.
func (user *user) applyMessageIDChanged(ctx context.Context, tx *ent.Tx, update *imap.MessageIDChanged) error {
	if err := user.forState(func(state *State) error {
		return state.updateMessageID(update.OldID, update.NewID)
	}); err != nil {
		return err
	}

	if err := user.remote.FinishMessageIDUpdate(update.OldID); err != nil {
		logrus.WithError(err).Error("Call to FinishMessageIDUpdate() failed")
	}

	if err := user.store.Update(update.OldID, update.NewID); err != nil {
		return err
	}

	return DBUpdateMessageID(ctx, tx, update.OldID, update.NewID)
}

func (user *user) setMessageMailboxes(ctx context.Context, tx *ent.Tx, messageID string, mboxIDs []string) error {
	curMailboxIDs, err := DBGetMessageMailboxIDs(ctx, tx.Client(), messageID)
	if err != nil {
		return err
	}

	for _, mboxID := range xslices.Filter(mboxIDs, func(mboxID string) bool { return !slices.Contains(curMailboxIDs, mboxID) }) {
		if _, err := user.applyMessagesAddedToMailbox(ctx, tx, mboxID, []string{messageID}); err != nil {
			return err
		}
	}

	for _, mboxID := range xslices.Filter(curMailboxIDs, func(mboxID string) bool { return !slices.Contains(mboxIDs, mboxID) }) {
		if err := user.applyMessagesRemovedFromMailbox(ctx, tx, mboxID, []string{messageID}); err != nil {
			return err
		}
	}

	return nil
}

// applyMessagesAddedToMailbox adds the messages to the given mailbox.
func (user *user) applyMessagesAddedToMailbox(ctx context.Context, tx *ent.Tx, mboxID string, messageIDs []string) (map[string]int, error) {
	if _, err := DBAddMessagesToMailbox(ctx, tx, messageIDs, mboxID); err != nil {
		return nil, err
	}

	messageUIDs, err := DBGetMessageUIDs(ctx, tx.Client(), mboxID, messageIDs)
	if err != nil {
		return nil, err
	}

	if err := user.forStateInMailbox(mboxID, func(other *State) error {
		for _, messageID := range messageIDs {
			if err := other.pushResponder(ctx, tx, newExists(messageID, messageUIDs[messageID])); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return messageUIDs, nil
}

// applyMessagesRemovedFromMailbox removes the messages from the given mailbox.
func (user *user) applyMessagesRemovedFromMailbox(ctx context.Context, tx *ent.Tx, mboxID string, messageIDs []string) error {
	if len(messageIDs) > 0 {
		if err := DBRemoveMessagesFromMailbox(ctx, tx, messageIDs, mboxID); err != nil {
			return err
		}
	}

	for _, messageID := range messageIDs {
		if err := user.forStateInMailboxWithMessage(mboxID, messageID, func(other *State) error {
			return other.pushResponder(ctx, tx, newExpunge(messageID, isClose(ctx)))
		}); err != nil {
			return err
		}
	}

	return nil
}

func (user *user) setMessageFlags(ctx context.Context, tx *ent.Tx, messageID string, seen, flagged bool) error {
	curFlags, err := DBGetMessageFlags(ctx, tx.Client(), []string{messageID})
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

func (user *user) addMessageFlags(ctx context.Context, tx *ent.Tx, messageID string, flag string) error {
	if err := DBAddMessageFlag(ctx, tx, []string{messageID}, flag); err != nil {
		return err
	}

	return user.forStateWithMessage(messageID, func(state *State) error {
		snapFlags, err := state.snap.getMessageFlags(messageID)
		if err != nil {
			return err
		}

		return state.pushResponder(ctx, tx, newFetch(messageID, snapFlags.Add(flag), isUID(ctx), isSilent(ctx)))
	})
}

func (user *user) removeMessageFlags(ctx context.Context, tx *ent.Tx, messageID string, flag string) error {
	if err := DBRemoveMessageFlag(ctx, tx, []string{messageID}, flag); err != nil {
		return err
	}

	return user.forStateWithMessage(messageID, func(state *State) error {
		snapFlags, err := state.snap.getMessageFlags(messageID)
		if err != nil {
			return err
		}

		return state.pushResponder(ctx, tx, newFetch(messageID, snapFlags.Remove(flag), isUID(ctx), isSilent(ctx)))
	})
}

func (user *user) applyMessageDeleted(ctx context.Context, tx *ent.Tx, update *imap.MessageDeleted) error {
	if err := DBMarkMessageAsDeleted(ctx, tx, update.MessageID); err != nil {
		return err
	}

	return user.forStateWithMessage(update.MessageID, func(state *State) error {
		return state.actionRemoveMessagesFromMailbox(ctx, tx, []string{update.MessageID}, state.snap.mboxID)
	})
}
