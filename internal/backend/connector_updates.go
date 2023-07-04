package backend

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/ProtonMail/gluon/db"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/ids"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/ProtonMail/gluon/rfc822"
	"github.com/bradenaw/juniper/parallel"
	"github.com/bradenaw/juniper/xslices"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

// apply an incoming update originating from the connector.
func (user *user) apply(ctx context.Context, update imap.Update) error {
	logrus.WithField("update", update).WithField("user-id", user.userID).Debug("Applying update")

	err := func() error {
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

		case *imap.MessageUpdated:
			return user.applyMessageUpdated(ctx, update)

		case *imap.UIDValidityBumped:
			return user.applyUIDValidityBumped(ctx, update)

		case *imap.Noop:
			return nil

		default:
			return fmt.Errorf("bad update")
		}
	}()

	update.Done(err)

	return err
}

// applyMailboxCreated applies a MailboxCreated update.
func (user *user) applyMailboxCreated(ctx context.Context, update *imap.MailboxCreated) error {
	if update.Mailbox.ID == ids.GluonInternalRecoveryMailboxRemoteID {
		return fmt.Errorf("attempting to create protected mailbox (recovery)")
	}

	if exists, err := db.ClientReadType(ctx, user.db, func(ctx context.Context, client db.ReadOnly) (bool, error) {
		return client.MailboxExistsWithRemoteID(ctx, update.Mailbox.ID)
	}); err != nil {
		return err
	} else if exists {
		return nil
	}

	uidValidity, err := user.uidValidityGenerator.Generate()
	if err != nil {
		return err
	}

	if err := user.imapLimits.CheckUIDValidity(uidValidity); err != nil {
		return err
	}

	return user.db.Write(ctx, func(ctx context.Context, tx db.Transaction) error {
		if mailboxCount, err := tx.GetMailboxCount(ctx); err != nil {
			return err
		} else if err := user.imapLimits.CheckMailBoxCount(mailboxCount); err != nil {
			return err
		}

		if _, err := tx.CreateMailbox(
			ctx,
			update.Mailbox.ID,
			strings.Join(update.Mailbox.Name, user.delimiter),
			update.Mailbox.Flags,
			update.Mailbox.PermanentFlags,
			update.Mailbox.Attributes,
			uidValidity,
		); err != nil {
			return err
		}

		return nil
	})
}

// applyMailboxDeleted applies a MailboxDeleted update.
func (user *user) applyMailboxDeleted(ctx context.Context, update *imap.MailboxDeleted) error {
	if update.MailboxID == ids.GluonInternalRecoveryMailboxRemoteID {
		return fmt.Errorf("attempting to delete protected mailbox (recovery)")
	}

	return userDBWrite(ctx, user, func(ctx context.Context, tx db.Transaction) ([]state.Update, error) {
		mailbox, err := tx.GetMailboxByRemoteID(ctx, update.MailboxID)
		if err != nil {
			if db.IsErrNotFound(err) {
				return nil, nil
			}

			return nil, err
		}

		if err := tx.DeleteMailboxWithRemoteID(ctx, update.MailboxID); err != nil {
			return nil, err
		}

		if _, err := tx.RemoveDeletedSubscriptionWithName(ctx, mailbox.Name); err != nil {
			return nil, err
		}

		return []state.Update{state.NewMailboxDeletedStateUpdate(mailbox.ID)}, nil
	})
}

// applyMailboxUpdated applies a MailboxUpdated update.
func (user *user) applyMailboxUpdated(ctx context.Context, update *imap.MailboxUpdated) error {
	if update.MailboxID == ids.GluonInternalRecoveryMailboxRemoteID {
		return fmt.Errorf("attempting to rename protected mailbox (recovery)")
	}

	return user.db.Write(ctx, func(ctx context.Context, tx db.Transaction) error {
		if exists, err := tx.MailboxExistsWithRemoteID(ctx, update.MailboxID); err != nil {
			return err
		} else if !exists {
			return nil
		}

		currentName, err := tx.GetMailboxNameWithRemoteID(ctx, update.MailboxID)
		if err != nil {
			return err
		}

		if currentName == strings.Join(update.MailboxName, user.delimiter) {
			return nil
		}

		return tx.RenameMailboxWithRemoteID(ctx, update.MailboxID, strings.Join(update.MailboxName, user.delimiter))
	})
}

// applyMailboxIDChanged applies a MailboxIDChanged update.
func (user *user) applyMailboxIDChanged(ctx context.Context, update *imap.MailboxIDChanged) error {
	if update.InternalID == user.recoveryMailboxID {
		return fmt.Errorf("attempting to change protected mailbox (recovery) remote ID")
	}

	return userDBWrite(ctx, user, func(ctx context.Context, tx db.Transaction) ([]state.Update, error) {
		if err := tx.UpdateRemoteMailboxID(ctx, update.InternalID, update.RemoteID); err != nil {
			return nil, err
		}

		return []state.Update{state.NewMailboxRemoteIDUpdateStateUpdate(update.InternalID, update.RemoteID)}, nil
	})
}

// applyMessagesCreated applies a MessagesCreated update.
func (user *user) applyMessagesCreated(ctx context.Context, update *imap.MessagesCreated) error {
	type DBRequestWithLiteral struct {
		db.CreateMessageReq
		reader io.Reader
	}

	// collect all unique messages to create
	messagesToCreate := make([]*DBRequestWithLiteral, 0, len(update.Messages))
	messagesToCreateFilter := make(map[imap.MessageID]imap.InternalMessageID, len(update.Messages)/2)
	messageForMBox := make(map[imap.InternalMailboxID][]db.MessageIDPair)
	mboxInternalIDMap := make(map[imap.MailboxID]imap.InternalMailboxID)

	err := userDBWrite(ctx, user, func(ctx context.Context, tx db.Transaction) ([]state.Update, error) {
		for _, message := range update.Messages {
			if slices.Contains(message.MailboxIDs, ids.GluonInternalRecoveryMailboxRemoteID) {
				logrus.Errorf("attempting to import messages into protected mailbox (recovery), skipping")
				continue
			}

			internalID, ok := messagesToCreateFilter[message.Message.ID]
			if !ok {
				messageID, err := tx.GetMessageIDFromRemoteID(ctx, message.Message.ID)
				if db.IsErrNotFound(err) {
					internalID = imap.NewInternalMessageID()

					literalReader, literalSize, err := rfc822.SetHeaderValueNoMemCopy(message.Literal, ids.InternalIDKey, internalID.String())
					if err != nil {
						return nil, fmt.Errorf("failed to set internal ID: %w", err)
					}

					request := &DBRequestWithLiteral{
						CreateMessageReq: db.CreateMessageReq{
							Message:     message.Message,
							LiteralSize: literalSize,
							Body:        message.ParsedMessage.Body,
							Structure:   message.ParsedMessage.Structure,
							Envelope:    message.ParsedMessage.Envelope,
							InternalID:  internalID,
						},
						reader: literalReader,
					}

					messagesToCreate = append(messagesToCreate, request)
					messagesToCreateFilter[message.Message.ID] = internalID
				} else if err == nil {
					internalID = messageID
				} else {
					return nil, err
				}
			}

			for _, mboxID := range message.MailboxIDs {
				v, ok := mboxInternalIDMap[mboxID]
				if !ok {
					internalMBoxID, err := tx.GetMailboxIDFromRemoteID(ctx, mboxID)
					if err != nil {
						// If a mailbox doesn't exist and we are allowed to skip move to next mailbox.
						if update.IgnoreUnknownMailboxIDs {
							logrus.WithField("MailboxID", mboxID.ShortID()).
								WithField("MessageID", message.Message.ID.ShortID()).
								Warn("Unknown Mailbox ID, skipping add to mailbox")
							continue
						}
						return nil, err
					}

					v = internalMBoxID
					mboxInternalIDMap[mboxID] = v
				}

				messageList, ok := messageForMBox[v]
				if !ok {
					messageList = []db.MessageIDPair{}
					messageForMBox[v] = messageList
				}

				if !slices.ContainsFunc(messageList, func(id db.MessageIDPair) bool { return id.InternalID == internalID }) {
					messageList = append(messageList, db.MessageIDPair{InternalID: internalID, RemoteID: message.Message.ID})
					messageForMBox[v] = messageList
				}
			}
		}

		if len(messagesToCreate) == 0 && len(messageForMBox) == 0 {
			return nil, nil
		}

		for _, chunk := range xslices.Chunk(messagesToCreate, db.ChunkLimit) {
			// Create messages in the store in parallel
			numStoreRoutines := runtime.NumCPU() / 4
			if numStoreRoutines < len(chunk) {
				numStoreRoutines = len(chunk)
			}
			if err := parallel.DoContext(ctx, numStoreRoutines, len(chunk), func(ctx context.Context, i int) error {
				msg := chunk[i]
				if err := user.store.SetUnchecked(msg.InternalID, msg.reader); err != nil {
					return fmt.Errorf("failed to store message literal: %w", err)
				}

				return nil
			}); err != nil {
				return nil, err
			}

			// Create message in the database
			if err := tx.CreateMessages(ctx, xslices.Map(chunk, func(req *DBRequestWithLiteral) *db.CreateMessageReq {
				return &req.CreateMessageReq
			})...); err != nil {
				return nil, err
			}
		}

		var stateUpdates []state.Update

		// Assign all the messages to the mailbox
		for mboxID, msgList := range messageForMBox {
			inMailbox, err := tx.MailboxFilterContains(ctx, mboxID, msgList)
			if err != nil {
				return nil, err
			}

			toAdd := xslices.Filter(msgList, func(id db.MessageIDPair) bool {
				return !slices.Contains(inMailbox, id.InternalID)
			})

			if len(toAdd) != 0 {
				_, update, err := user.applyMessagesAddedToMailbox(ctx, tx, mboxID, toAdd)
				if err != nil {
					return nil, err
				}

				stateUpdates = append(stateUpdates, update)
			}
		}

		return stateUpdates, nil
	})

	// Clean up cache messages that were created if the transaction failed.
	if err != nil {
		for _, message := range messagesToCreate {
			if err := user.store.DeleteUnchecked(message.InternalID); err != nil {
				if !os.IsNotExist(err) {
					logrus.WithError(err).Errorf("Failed to delete cache message %v after failed transaction", message.InternalID)
				}
			}
		}
	}

	return err
}

// applyMessageMailboxesUpdated applies a MessageMailboxesUpdated update.
func (user *user) applyMessageMailboxesUpdated(ctx context.Context, update *imap.MessageMailboxesUpdated) error {
	if slices.Contains(update.MailboxIDs, ids.GluonInternalRecoveryMailboxRemoteID) {
		return fmt.Errorf("attempting to move messages into protected mailbox (recovery)")
	}

	if exists, err := db.ClientReadType(ctx, user.db, func(ctx context.Context, client db.ReadOnly) (bool, error) {
		return client.MessageExistsWithRemoteID(ctx, update.MessageID)
	}); err != nil {
		return err
	} else if !exists {
		return state.ErrNoSuchMessage
	}

	return userDBWrite(ctx, user, func(ctx context.Context, tx db.Transaction) ([]state.Update, error) {
		internalMsgID, err := tx.GetMessageIDFromRemoteID(ctx, update.MessageID)
		if err != nil {
			if db.IsErrNotFound(err) {
				return nil, state.ErrNoSuchMessage
			}
			return nil, err
		}

		internalMBoxIDs, err := tx.MailboxTranslateRemoteIDs(ctx, update.MailboxIDs)
		if err != nil {
			return nil, err
		}

		var stateUpdates []state.Update

		mboxUpdates, err := user.setMessageMailboxes(ctx, tx, db.MessageIDPair{
			InternalID: internalMsgID,
			RemoteID:   update.MessageID,
		}, internalMBoxIDs)
		if err != nil {
			return nil, err
		}

		stateUpdates = append(stateUpdates, mboxUpdates...)

		flagUpdates, err := user.setMessageFlags(ctx, tx, internalMsgID, update.Flags)
		if err != nil {
			return nil, err
		}

		return append(stateUpdates, flagUpdates...), nil
	})
}

// applyMessageFlagsUpdated applies a MessageFlagsUpdated update.
func (user *user) applyMessageFlagsUpdated(ctx context.Context, update *imap.MessageFlagsUpdated) error {
	if exists, err := db.ClientReadType(ctx, user.db, func(ctx context.Context, client db.ReadOnly) (bool, error) {
		return client.MessageExistsWithRemoteID(ctx, update.MessageID)
	}); err != nil {
		return err
	} else if !exists {
		return state.ErrNoSuchMessage
	}

	return userDBWrite(ctx, user, func(ctx context.Context, tx db.Transaction) ([]state.Update, error) {
		internalMsgID, err := tx.GetMessageIDFromRemoteID(ctx, update.MessageID)

		if err != nil {
			if db.IsErrNotFound(err) {
				return nil, state.ErrNoSuchMessage
			}
			return nil, err
		}

		return user.setMessageFlags(ctx, tx, internalMsgID, update.Flags)
	})
}

// applyMessageIDChanged applies a MessageIDChanged update.
func (user *user) applyMessageIDChanged(ctx context.Context, update *imap.MessageIDChanged) error {
	if err := user.db.Write(ctx, func(ctx context.Context, tx db.Transaction) error {
		return tx.UpdateRemoteMessageID(ctx, update.InternalID, update.RemoteID)
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

func (user *user) setMessageMailboxes(ctx context.Context, tx db.Transaction, messageID db.MessageIDPair, mboxIDs []imap.InternalMailboxID) ([]state.Update, error) {
	curMailboxIDs, err := tx.GetMessageMailboxIDs(ctx, messageID.InternalID)
	if err != nil {
		return nil, err
	}

	var updates []state.Update

	for _, mboxID := range xslices.Filter(mboxIDs, func(mboxID imap.InternalMailboxID) bool { return !slices.Contains(curMailboxIDs, mboxID) }) {
		_, update, err := user.applyMessagesAddedToMailbox(ctx, tx, mboxID, []db.MessageIDPair{messageID})
		if err != nil {
			return nil, err
		}

		updates = append(updates, update)
	}

	for _, mboxID := range xslices.Filter(curMailboxIDs, func(mboxID imap.InternalMailboxID) bool { return !slices.Contains(mboxIDs, mboxID) }) {
		update, err := user.applyMessagesRemovedFromMailbox(ctx, tx, mboxID, []imap.InternalMessageID{messageID.InternalID})
		if err != nil {
			return nil, err
		}

		updates = append(updates, update...)
	}

	return updates, nil
}

// applyMessagesAddedToMailbox adds the messages to the given mailbox.
func (user *user) applyMessagesAddedToMailbox(ctx context.Context, tx db.Transaction, mboxID imap.InternalMailboxID, messageIDs []db.MessageIDPair) ([]db.UIDWithFlags, state.Update, error) {
	return state.AddMessagesToMailbox(ctx, tx, mboxID, messageIDs, nil, user.imapLimits)
}

// applyMessagesRemovedFromMailbox removes the messages from the given mailbox.
func (user *user) applyMessagesRemovedFromMailbox(ctx context.Context, tx db.Transaction, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) ([]state.Update, error) {
	return state.RemoveMessagesFromMailbox(ctx, tx, mboxID, messageIDs)
}

func (user *user) setMessageFlags(ctx context.Context, tx db.Transaction, messageID imap.InternalMessageID, flags imap.FlagSet) ([]state.Update, error) {
	var stateUpdates []state.Update

	curFlags, err := tx.GetMessagesFlags(ctx, []imap.InternalMessageID{messageID})
	if err != nil {
		return nil, err
	}

	flagSet := curFlags[0].FlagSet

	for _, v := range flagSet.ToSliceUnsorted() {
		if !flags.Contains(v) {
			update, err := user.removeMessageFlags(ctx, tx, messageID, v)
			if err != nil {
				return nil, err
			}

			stateUpdates = append(stateUpdates, update)
		}
	}

	for _, v := range flags.ToSliceUnsorted() {
		if !flagSet.Contains(v) {
			update, err := user.addMessageFlags(ctx, tx, messageID, v)
			if err != nil {
				return nil, err
			}

			stateUpdates = append(stateUpdates, update)
		}
	}

	return stateUpdates, nil
}

func (user *user) addMessageFlags(ctx context.Context, tx db.Transaction, messageID imap.InternalMessageID, flag string) (state.Update, error) {
	if err := tx.AddFlagToMessages(ctx, []imap.InternalMessageID{messageID}, flag); err != nil {
		return nil, err
	}

	return state.NewRemoteAddMessageFlagsStateUpdate(messageID, flag), nil
}

func (user *user) removeMessageFlags(ctx context.Context, tx db.Transaction, messageID imap.InternalMessageID, flag string) (state.Update, error) {
	if err := tx.RemoveFlagFromMessages(ctx, []imap.InternalMessageID{messageID}, flag); err != nil {
		return nil, err
	}

	return state.NewRemoteRemoveMessageFlagsStateUpdate(messageID, flag), nil
}

func (user *user) applyMessageDeleted(ctx context.Context, update *imap.MessageDeleted) error {
	if err := userDBWrite(ctx, user, func(ctx context.Context, tx db.Transaction) ([]state.Update, error) {
		if err := tx.MarkMessageAsDeletedWithRemoteID(ctx, update.MessageID); err != nil {
			if db.IsErrNotFound(err) {
				return nil, nil
			}

			return nil, err
		}

		internalMessageID, err := tx.GetMessageIDFromRemoteID(ctx, update.MessageID)
		if err != nil {
			if db.IsErrNotFound(err) {
				return nil, nil
			}

			return nil, err
		}

		mailboxes, err := tx.GetMessageMailboxIDs(ctx, internalMessageID)
		if err != nil {
			return nil, err
		}

		messageIDs := []imap.InternalMessageID{internalMessageID}

		var stateUpdates []state.Update

		for _, mailbox := range mailboxes {
			updates, err := state.RemoveMessagesFromMailbox(ctx, tx, mailbox, messageIDs)
			if err != nil {
				return nil, err
			}

			stateUpdates = append(stateUpdates, updates...)
		}

		return stateUpdates, nil
	}); err != nil {
		return err
	}

	return nil
}

func (user *user) applyMessageUpdated(ctx context.Context, update *imap.MessageUpdated) error {
	log := logrus.WithField("message updated", update.Message.ID.ShortID())

	internalMessageID, err := db.ClientReadType(ctx, user.db, func(ctx context.Context, client db.ReadOnly) (imap.InternalMessageID, error) {
		return client.GetMessageIDFromRemoteID(ctx, update.Message.ID)
	})
	if db.IsErrNotFound(err) {
		if update.AllowCreate {
			log.Warn("Message not found, creating it instead")

			return user.applyMessagesCreated(ctx, imap.NewMessagesCreated(true, &imap.MessageCreated{
				Message:       update.Message,
				Literal:       update.Literal,
				MailboxIDs:    update.MailboxIDs,
				ParsedMessage: update.ParsedMessage,
			}))
		} else {
			log.Warn("Message not found, skipping update")
			return nil
		}
	} else if err != nil {
		return err
	}

	return userDBWrite(ctx, user, func(ctx context.Context, tx db.Transaction) ([]state.Update, error) {
		// compare and see if the literal has changed.
		onDiskLiteral, err := user.store.Get(internalMessageID)
		if err != nil {
			logrus.Debugf("failed to retrieve literal from cache: %v", err)
		}

		updateLiteral := update.Literal
		if id, err := rfc822.GetHeaderValue(updateLiteral, ids.InternalIDKey); err == nil {
			if len(id) == 0 {
				l, err := rfc822.SetHeaderValue(updateLiteral, ids.InternalIDKey, internalMessageID.String())
				if err != nil {
					log.WithError(err).Debug("failed to set header key, using update literal")
				} else {
					updateLiteral = l
				}
			}
		} else {
			log.Debug("Failed to get header value from literal, using update literal")
		}

		var stateUpdates []state.Update

		if bytes.Equal(onDiskLiteral, updateLiteral) {
			log.Debug("Message not updated as there are no changes to literals, assigning mailboxes only")

			targetMailboxes := make([]imap.InternalMailboxID, 0, len(update.MailboxIDs))

			for _, mbox := range update.MailboxIDs {
				internalMBoxID, err := tx.GetMailboxIDFromRemoteID(ctx, mbox)
				if err != nil {
					return nil, err
				}

				targetMailboxes = append(targetMailboxes, internalMBoxID)
			}

			flagUpdates, err := user.setMessageFlags(ctx, tx, internalMessageID, update.Message.Flags)
			if err != nil {
				return nil, err
			}

			stateUpdates = append(stateUpdates, flagUpdates...)

			mboxUpdates, err := user.setMessageMailboxes(ctx, tx, db.MessageIDPair{
				InternalID: internalMessageID,
				RemoteID:   update.Message.ID,
			}, targetMailboxes)
			if err != nil {
				return nil, err
			}

			return append(stateUpdates, mboxUpdates...), nil
		} else {
			log.Debug("Message has new literal, applying update")

			{
				// delete the message and remove from the mailboxes.
				mailboxes, err := tx.GetMessageMailboxIDs(ctx, internalMessageID)
				if err != nil {
					return nil, err
				}

				messageIDs := []imap.InternalMessageID{internalMessageID}

				for _, mailbox := range mailboxes {
					updates, err := state.RemoveMessagesFromMailbox(ctx, tx, mailbox, messageIDs)
					if err != nil {
						return nil, err
					}

					stateUpdates = append(stateUpdates, updates...)
				}

				// We need change the old remote id as it will break our table constraint otherwise and everything
				// will silently fail.
				if err := tx.MarkMessageAsDeletedAndAssignRandomRemoteID(ctx, internalMessageID); err != nil {
					return nil, err
				}
			}
			// create new entry
			{
				newInternalID := imap.NewInternalMessageID()

				literalReader, literalSize, err := rfc822.SetHeaderValueNoMemCopy(update.Literal, ids.InternalIDKey, newInternalID.String())
				if err != nil {
					return nil, fmt.Errorf("failed to set internal ID: %w", err)
				}

				request := &db.CreateMessageReq{
					Message:     update.Message,
					LiteralSize: literalSize,
					Body:        update.ParsedMessage.Body,
					Structure:   update.ParsedMessage.Structure,
					Envelope:    update.ParsedMessage.Envelope,
					InternalID:  newInternalID,
				}

				if err := tx.CreateMessages(ctx, request); err != nil {
					return nil, err
				}

				if err := user.store.Set(newInternalID, literalReader); err != nil {
					return nil, err
				}

				for _, mbox := range update.MailboxIDs {
					internalMBoxID, err := tx.GetMailboxIDFromRemoteID(ctx, mbox)
					if err != nil {
						return nil, err
					}

					_, update, err := state.AddMessagesToMailbox(
						ctx,
						tx,
						internalMBoxID,
						[]db.MessageIDPair{{InternalID: newInternalID, RemoteID: update.Message.ID}},
						nil,
						user.imapLimits,
					)
					if err != nil {
						return nil, err
					}

					stateUpdates = append(stateUpdates, update)
				}
			}
		}

		return stateUpdates, nil
	})
}

// applyUIDValidityBumped applies a UIDValidityBumped event to the user.
func (user *user) applyUIDValidityBumped(ctx context.Context, update *imap.UIDValidityBumped) error {
	if err := user.db.Write(ctx, func(ctx context.Context, tx db.Transaction) error {
		mailboxes, err := tx.GetAllMailboxesWithAttr(ctx)
		if err != nil {
			return err
		}

		for _, mailbox := range mailboxes {
			uidValidity, err := user.uidValidityGenerator.Generate()
			if err != nil {
				return err
			}

			if err := tx.SetMailboxUIDValidity(ctx, mailbox.ID, uidValidity); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	user.queueStateUpdate(state.NewUIDValidityBumpedStateUpdate())

	return nil
}

func userDBWrite(ctx context.Context, user *user, fn func(context.Context, db.Transaction) ([]state.Update, error)) error {
	var updates []state.Update

	if err := user.db.Write(ctx, func(ctx context.Context, tx db.Transaction) error {
		up, err := fn(ctx, tx)
		updates = up
		return err
	}); err != nil {
		return err
	}

	if len(updates) != 0 {
		user.queueStateUpdate(updates...)
	}

	return nil
}

func userDBWriteResult[T any](ctx context.Context, user *user, fn func(context.Context, db.Transaction) ([]state.Update, T, error)) (T, error) {
	var updates []state.Update

	result, err := db.ClientWriteType(ctx, user.db, func(ctx context.Context, tx db.Transaction) (T, error) {
		up, val, err := fn(ctx, tx)
		updates = up
		return val, err
	})
	if err != nil {
		var t T
		return t, err
	}

	// need to create a separate transaction for the state updates so that import changes get written first.
	if len(updates) != 0 {
		user.queueStateUpdate(updates...)
	}

	return result, nil
}
