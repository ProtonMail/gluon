package backend

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"runtime"
	"strings"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db"
	"github.com/ProtonMail/gluon/internal/db/ent"
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

	if exists, err := db.ReadResult(ctx, user.db, func(ctx context.Context, client *ent.Client) (bool, error) {
		return db.MailboxExistsWithRemoteID(ctx, client, update.Mailbox.ID)
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

	return user.db.Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		if mailboxCount, err := db.GetMailboxCount(ctx, tx.Client()); err != nil {
			return err
		} else if err := user.imapLimits.CheckMailBoxCount(mailboxCount); err != nil {
			return err
		}

		if _, err := db.CreateMailbox(
			ctx,
			tx,
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
		return db.DeleteMailboxWithRemoteID(ctx, tx, update.MailboxID)
	}); err != nil {
		return err
	}

	user.queueStateUpdate(state.NewMailboxDeletedStateUpdate(internalMailboxID))

	return nil
}

// applyMailboxUpdated applies a MailboxUpdated update.
func (user *user) applyMailboxUpdated(ctx context.Context, update *imap.MailboxUpdated) error {
	if update.MailboxID == ids.GluonInternalRecoveryMailboxRemoteID {
		return fmt.Errorf("attempting to rename protected mailbox (recovery)")
	}

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
	if update.InternalID == user.recoveryMailboxID {
		return fmt.Errorf("attempting to change protected mailbox (recovery) remote ID")
	}

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
	type DBRequestWithLiteral struct {
		db.CreateMessageReq
		reader io.Reader
	}

	// collect all unique messages to create
	messagesToCreate := make([]*DBRequestWithLiteral, 0, len(update.Messages))
	messagesToCreateFilter := make(map[imap.MessageID]imap.InternalMessageID, len(update.Messages)/2)
	messageForMBox := make(map[imap.InternalMailboxID][]imap.InternalMessageID)
	mboxInternalIDMap := make(map[imap.MailboxID]imap.InternalMailboxID)

	if err := user.db.Read(ctx, func(ctx context.Context, client *ent.Client) error {
		for _, message := range update.Messages {
			if slices.Contains(message.MailboxIDs, ids.GluonInternalRecoveryMailboxRemoteID) {
				logrus.Errorf("attempting to import messages into protected mailbox (recovery), skipping")
				continue
			}

			internalID, ok := messagesToCreateFilter[message.Message.ID]
			if !ok {
				messageID, err := db.GetMessageIDFromRemoteID(ctx, client, message.Message.ID)
				if ent.IsNotFound(err) {
					internalID = imap.NewInternalMessageID()

					literalReader, literalSize, err := rfc822.SetHeaderValueNoMemCopy(message.Literal, ids.InternalIDKey, internalID.String())
					if err != nil {
						return fmt.Errorf("failed to set internal ID: %w", err)
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
					return err
				}
			}

			for _, mboxID := range message.MailboxIDs {
				v, ok := mboxInternalIDMap[mboxID]
				if !ok {
					internalMBoxID, err := db.GetMailboxIDWithRemoteID(ctx, client, mboxID)
					if err != nil {
						// If a mailbox doesn't exist and we are allowed to skip move to next mailbox.
						if update.IgnoreUnknownMailboxIDs {
							logrus.WithField("MailboxID", mboxID.ShortID()).
								WithField("MessageID", message.Message.ID.ShortID()).
								Warn("Unknown Mailbox ID, skipping add to mailbox")
							continue
						}
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
	if err := user.db.Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
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
				return err
			}

			// Create message in the database
			if _, err := db.CreateMessages(ctx, tx, xslices.Map(chunk, func(req *DBRequestWithLiteral) *db.CreateMessageReq {
				return &req.CreateMessageReq
			})...); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return user.db.Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		// Assign all the messages to the mailbox
		for mboxID, msgList := range messageForMBox {
			inMailbox, err := db.FilterMailboxContainsInternalID(ctx, tx.Client(), mboxID, msgList)
			if err != nil {
				return err
			}

			toAdd := xslices.Filter(msgList, func(id imap.InternalMessageID) bool {
				return !slices.Contains(inMailbox, id)
			})

			if len(toAdd) != 0 {
				if _, err := user.applyMessagesAddedToMailbox(ctx, tx, mboxID, toAdd); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// applyMessageMailboxesUpdated applies a MessageMailboxesUpdated update.
func (user *user) applyMessageMailboxesUpdated(ctx context.Context, update *imap.MessageMailboxesUpdated) error {
	if slices.Contains(update.MailboxIDs, ids.GluonInternalRecoveryMailboxRemoteID) {
		return fmt.Errorf("attempting to move messages into protected mailbox (recovery)")
	}

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

		if err := user.setMessageFlags(ctx, tx, result.InternalMsgID, update.Seen, update.Flagged, update.Draft); err != nil {
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
		if err := user.setMessageFlags(ctx, tx, internalMsgID, update.Seen, update.Flagged, update.Draft); err != nil {
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
	messageUIDs, update, err := state.AddMessagesToMailbox(ctx, tx, mboxID, messageIDs, nil, user.imapLimits)
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

func (user *user) setMessageFlags(ctx context.Context, tx *ent.Tx, messageID imap.InternalMessageID, seen, flagged, draft bool) error {
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

	if draft && !flagSet.ContainsUnchecked(imap.FlagDraftLowerCase) {
		if err := user.addMessageFlags(ctx, tx, messageID, imap.FlagDraft); err != nil {
			return err
		}
	} else if !draft && flagSet.ContainsUnchecked(imap.FlagDraftLowerCase) {
		if err := user.removeMessageFlags(ctx, tx, messageID, imap.FlagDraft); err != nil {
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
			if ent.IsNotFound(err) {
				return nil
			}

			return err
		}

		internalMessageID, err := db.GetMessageIDFromRemoteID(ctx, tx.Client(), update.MessageID)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil
			}

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

func (user *user) applyMessageUpdated(ctx context.Context, update *imap.MessageUpdated) error {
	log := logrus.WithField("message updated", update.Message.ID.ShortID())

	if err := user.db.Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		internalMessageID, err := db.GetMessageIDFromRemoteID(ctx, tx.Client(), update.Message.ID)
		if err != nil {
			if ent.IsNotFound(err) {
				log.Debugf("Message %v no longer exists, skipping", update.Message.ID.ShortID())
				return nil
			}
			return err
		}

		// compare and see if the literal has changed.
		onDiskLiteral, err := user.store.Get(internalMessageID)
		if err != nil {
			return fmt.Errorf("failed to retrieve literal from cache: %w", err)
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

		if bytes.Equal(onDiskLiteral, updateLiteral) {
			log.Debug("Message not updated as there are no changes to literals, assigning mailboxes only")

			targetMailboxes := make([]imap.InternalMailboxID, 0, len(update.MailboxIDs))

			for _, mbox := range update.MailboxIDs {
				internalMBoxID, err := db.GetMailboxIDFromRemoteID(ctx, tx.Client(), mbox)
				if err != nil {
					return err
				}

				targetMailboxes = append(targetMailboxes, internalMBoxID)
			}

			return user.setMessageMailboxes(ctx, tx, internalMessageID, targetMailboxes)
		} else {
			log.Debug("Message has new literal, applying update")

			var stateUpdates []state.Update
			{
				// delete the message and remove from the mailboxes.
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

				if err := db.DeleteMessages(ctx, tx, internalMessageID); err != nil {
					return err
				}
			}
			// create new entry
			{
				newInternalID := imap.NewInternalMessageID()

				literalReader, literalSize, err := rfc822.SetHeaderValueNoMemCopy(update.Literal, ids.InternalIDKey, newInternalID.String())
				if err != nil {
					return fmt.Errorf("failed to set internal ID: %w", err)
				}

				request := &db.CreateMessageReq{
					Message:     update.Message,
					LiteralSize: literalSize,
					Body:        update.ParsedMessage.Body,
					Structure:   update.ParsedMessage.Structure,
					Envelope:    update.ParsedMessage.Envelope,
					InternalID:  newInternalID,
				}

				if _, err := db.CreateMessages(ctx, tx, request); err != nil {
					return err
				}

				if err := user.store.Set(newInternalID, literalReader); err != nil {
					return err
				}

				for _, mbox := range update.MailboxIDs {
					internalMBoxID, err := db.GetMailboxIDFromRemoteID(ctx, tx.Client(), mbox)
					if err != nil {
						return err
					}

					_, update, err := state.AddMessagesToMailbox(ctx, tx, internalMBoxID, []imap.InternalMessageID{newInternalID}, nil, user.imapLimits)
					if err != nil {
						return err
					}

					stateUpdates = append(stateUpdates, update)
				}
			}

			if len(stateUpdates) != 0 {
				user.queueStateUpdate(stateUpdates...)
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

// applyUIDValidityBumped applies a UIDValidityBumped event to the user.
func (user *user) applyUIDValidityBumped(ctx context.Context, update *imap.UIDValidityBumped) error {
	if err := user.db.Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		mailboxes, err := db.GetAllMailboxes(ctx, tx.Client())
		if err != nil {
			return err
		}

		for _, mailbox := range mailboxes {
			uidValidity, err := user.uidValidityGenerator.Generate()
			if err != nil {
				return err
			}

			if _, err := mailbox.Update().SetUIDValidity(uidValidity).Save(ctx); err != nil {
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
