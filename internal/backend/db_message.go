package backend

import (
	"context"

	"github.com/ProtonMail/gluon/internal/backend/ent/messageflag"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/backend/ent"
	"github.com/ProtonMail/gluon/internal/backend/ent/mailbox"
	"github.com/ProtonMail/gluon/internal/backend/ent/message"
	"github.com/ProtonMail/gluon/internal/backend/ent/uid"
	"github.com/bradenaw/juniper/xslices"
)

type DBCreateMessageReq struct {
	message    imap.Message
	internalID imap.InternalMessageID
	literal    []byte
	body       string
	structure  string
	envelope   string
}

func DBCreateMessages(ctx context.Context, tx *ent.Tx, reqs ...*DBCreateMessageReq) ([]*ent.Message, error) {
	flags := make(map[imap.InternalMessageID][]*ent.MessageFlag)

	for _, req := range reqs {
		builders := xslices.Map(req.message.Flags.ToSlice(), func(flag string) *ent.MessageFlagCreate {
			return tx.MessageFlag.Create().SetValue(flag)
		})

		entFlags, err := tx.MessageFlag.CreateBulk(builders...).Save(ctx)
		if err != nil {
			return nil, err
		}

		flags[req.internalID] = entFlags
	}

	return tx.Message.CreateBulk(xslices.Map(reqs, func(req *DBCreateMessageReq) *ent.MessageCreate {
		msgCreate := tx.Message.Create().
			SetMessageID(req.internalID).
			SetDate(req.message.Date).
			SetBody(req.body).
			SetBodyStructure(req.structure).
			SetEnvelope(req.envelope).
			SetSize(len(req.literal)).
			AddFlags(flags[req.internalID]...)

		if len(req.message.ID) != 0 {
			msgCreate = msgCreate.SetRemoteID(req.message.ID)
		}

		return msgCreate
	})...).Save(ctx)
}

func DBAddMessagesToMailbox(ctx context.Context, tx *ent.Tx, messageIDs []imap.InternalMessageID, mboxID imap.InternalMailboxID) (map[imap.InternalMessageID]int, error) {
	messageUIDs := make(map[imap.InternalMessageID]int)

	mbox, err := tx.Mailbox.Query().Where(mailbox.MailboxID(mboxID)).Only(ctx)
	if err != nil {
		return nil, err
	}

	messages, err := DBGetMessages(ctx, tx.Client(), messageIDs...)
	if err != nil {
		return nil, err
	}

	var builders []*ent.UIDCreate

	for idx, messageID := range messageIDs {
		messageUIDs[messageID] = mbox.UIDNext + idx

		builders = append(builders, tx.UID.Create().
			SetMailbox(mbox).
			SetMessage(messages[messageID]).
			SetUID(messageUIDs[messageID]),
		)
	}

	if _, err := tx.UID.CreateBulk(builders...).Save(ctx); err != nil {
		return nil, err
	}

	if err := DBBumpMailboxUIDNext(ctx, tx, mbox, len(messageIDs)); err != nil {
		return nil, err
	}

	return messageUIDs, nil
}

func DBBumpMailboxUIDsForMessage(ctx context.Context, tx *ent.Tx, messageIDs []imap.InternalMessageID, mboxID imap.InternalMailboxID) (map[imap.InternalMessageID]int, error) {
	messageUIDs := make(map[imap.InternalMessageID]int)

	mbox, err := tx.Mailbox.Query().Where(mailbox.MailboxID(mboxID)).Only(ctx)
	if err != nil {
		return nil, err
	}

	var builders []*ent.UIDUpdate

	for idx, messageID := range messageIDs {
		uidNext := mbox.UIDNext + idx
		messageUIDs[messageID] = uidNext

		builders = append(builders, tx.UID.Update().
			SetUID(uidNext).
			Where(uid.HasMessageWith(message.MessageID(messageID))),
		)
	}

	for _, builder := range builders {
		if _, err := builder.Save(ctx); err != nil {
			return nil, err
		}
	}

	if err := DBBumpMailboxUIDNext(ctx, tx, mbox, len(messageIDs)); err != nil {
		return nil, err
	}

	return messageUIDs, nil
}

func DBRemoveMessagesFromMailbox(ctx context.Context, tx *ent.Tx, messageIDs []imap.InternalMessageID, mboxID imap.InternalMailboxID) error {
	if _, err := tx.UID.Delete().
		Where(
			uid.HasMailboxWith(mailbox.MailboxID(mboxID)),
			uid.HasMessageWith(message.MessageIDIn(messageIDs...)),
		).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func DBMessageExists(ctx context.Context, client *ent.Client, messageID imap.InternalMessageID) (bool, error) {
	count, err := client.Message.Query().Where(message.MessageID(messageID)).Count(ctx)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func DBMessageExistsWithRemoteID(ctx context.Context, client *ent.Client, messageID imap.MessageID) (bool, error) {
	count, err := client.Message.Query().Where(message.RemoteID(messageID)).Count(ctx)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func DBGetMessage(ctx context.Context, client *ent.Client, messageID imap.InternalMessageID) (*ent.Message, error) {
	return client.Message.Query().Where(message.MessageID(messageID)).Only(ctx)
}

func DBGetMessages(ctx context.Context, client *ent.Client, messageIDs ...imap.InternalMessageID) (map[imap.InternalMessageID]*ent.Message, error) {
	messages, err := client.Message.Query().Where(message.MessageIDIn(messageIDs...)).All(ctx)
	if err != nil {
		return nil, err
	}

	res := make(map[imap.InternalMessageID]*ent.Message)

	for _, message := range messages {
		res[message.MessageID] = message
	}

	return res, nil
}

func DBGetRemoteMessageID(ctx context.Context, client *ent.Client, internalID imap.InternalMessageID) (imap.MessageID, error) {
	message, err := client.Message.Query().Where(message.MessageID(internalID)).Select(message.FieldRemoteID).Only(ctx)
	if err != nil {
		return "", err
	}

	return message.RemoteID, nil
}

func DBGetMessageMailboxIDs(ctx context.Context, client *ent.Client, messageID imap.InternalMessageID) ([]imap.InternalMailboxID, error) {
	messageUIDs, err := client.UID.Query().
		Where(uid.HasMessageWith(message.MessageID(messageID))).
		WithMailbox(func(query *ent.MailboxQuery) {
			query.Select(mailbox.FieldMailboxID)
		}).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return xslices.Map(messageUIDs, func(message *ent.UID) imap.InternalMailboxID {
		return message.Edges.Mailbox.MailboxID
	}), nil
}

func DBGetMessageUIDs(ctx context.Context, client *ent.Client, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) (map[imap.InternalMessageID]int, error) {
	messageUIDs, err := client.UID.Query().
		Where(
			uid.HasMailboxWith(mailbox.MailboxID(mboxID)),
			uid.HasMessageWith(message.MessageIDIn(messageIDs...)),
		).
		WithMessage(func(query *ent.MessageQuery) {
			query.Select(message.FieldMessageID)
		}).
		All(ctx)
	if err != nil {
		return nil, err
	}

	res := make(map[imap.InternalMessageID]int)

	for _, messageUID := range messageUIDs {
		res[messageUID.Edges.Message.MessageID] = messageUID.UID
	}

	return res, nil
}

// DBGetMessageFlags returns the flags of the given messages.
// It does not include per-mailbox flags (\Deleted, \Recent)!
func DBGetMessageFlags(ctx context.Context, client *ent.Client, messageIDs []imap.InternalMessageID) (map[imap.InternalMessageID]imap.FlagSet, error) {
	messages, err := client.Message.Query().
		Where(message.MessageIDIn(messageIDs...)).
		WithFlags(func(query *ent.MessageFlagQuery) {
			query.Select(messageflag.FieldValue)
		}).
		All(ctx)
	if err != nil {
		return nil, err
	}

	curFlags := make(map[imap.InternalMessageID]imap.FlagSet)

	for _, message := range messages {
		curFlags[message.MessageID] = imap.NewFlagSetFromSlice(xslices.Map(message.Edges.Flags, func(flag *ent.MessageFlag) string {
			return flag.Value
		}))
	}

	return curFlags, nil
}

func DBGetMessageDeleted(ctx context.Context, client *ent.Client, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) (map[imap.InternalMessageID]bool, error) {
	messageUIDs, err := client.UID.Query().
		Where(
			uid.HasMailboxWith(mailbox.MailboxID(mboxID)),
			uid.HasMessageWith(message.MessageIDIn(messageIDs...)),
		).
		WithMessage(func(query *ent.MessageQuery) {
			query.Select(message.FieldMessageID)
		}).
		All(ctx)
	if err != nil {
		return nil, err
	}

	res := make(map[imap.InternalMessageID]bool)

	for _, messageUID := range messageUIDs {
		res[messageUID.Edges.Message.MessageID] = messageUID.Deleted
	}

	return res, nil
}

func DBAddMessageFlag(ctx context.Context, tx *ent.Tx, messageIDs []imap.InternalMessageID, addFlag string) error {
	messages, err := tx.Message.Query().
		Where(message.MessageIDIn(messageIDs...)).
		All(ctx)
	if err != nil {
		return err
	}

	builders := xslices.Map(messages, func(*ent.Message) *ent.MessageFlagCreate {
		return tx.MessageFlag.Create().SetValue(addFlag)
	})

	flags, err := tx.MessageFlag.CreateBulk(builders...).Save(ctx)
	if err != nil {
		return err
	}

	for idx, message := range messages {
		if _, err := message.Update().AddFlags(flags[idx]).Save(ctx); err != nil {
			return err
		}
	}

	return nil
}

func DBRemoveMessageFlag(ctx context.Context, tx *ent.Tx, messageIDs []imap.InternalMessageID, remFlag string) error {
	messages, err := tx.Message.Query().
		Where(message.MessageIDIn(messageIDs...)).
		WithFlags().
		All(ctx)
	if err != nil {
		return err
	}

	flags := xslices.Map(messages, func(message *ent.Message) *ent.MessageFlag {
		return message.Edges.Flags[xslices.IndexFunc(message.Edges.Flags, func(flag *ent.MessageFlag) bool {
			return imap.NewFlagSet(remFlag).Contains(flag.Value)
		})]
	})

	for idx, message := range messages {
		if _, err := message.Update().RemoveFlags(flags[idx]).Save(ctx); err != nil {
			return err
		}
	}

	return nil
}

func DBSetMessageFlags(ctx context.Context, tx *ent.Tx, messageIDs []imap.InternalMessageID, setFlags imap.FlagSet) error {
	messages, err := tx.Message.Query().
		Where(message.MessageIDIn(messageIDs...)).
		WithFlags().
		All(ctx)
	if err != nil {
		return err
	}

	for _, message := range messages {
		curFlagSet := imap.NewFlagSetFromSlice(xslices.Map(message.Edges.Flags, func(flag *ent.MessageFlag) string {
			return flag.Value
		}))

		addFlags := xslices.Filter(setFlags.ToSlice(), func(flag string) bool {
			return !curFlagSet.Contains(flag)
		})

		builders := xslices.Map(addFlags, func(flag string) *ent.MessageFlagCreate {
			return tx.MessageFlag.Create().SetValue(flag)
		})

		addEntFlags, err := tx.MessageFlag.CreateBulk(builders...).Save(ctx)
		if err != nil {
			return err
		}

		remEntFlags := xslices.Filter(message.Edges.Flags, func(flag *ent.MessageFlag) bool {
			return !setFlags.Contains(flag.Value)
		})

		if _, err := message.Update().
			AddFlags(addEntFlags...).
			RemoveFlags(remEntFlags...).
			Save(ctx); err != nil {
			return err
		}
	}

	return nil
}

func DBSetDeletedFlag(ctx context.Context, tx *ent.Tx, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID, deleted bool) error {
	if _, err := tx.UID.Update().
		Where(
			uid.HasMailboxWith(mailbox.MailboxID(mboxID)),
			uid.HasMessageWith(message.MessageIDIn(messageIDs...)),
		).
		SetDeleted(deleted).
		Save(ctx); err != nil {
		return err
	}

	return nil
}

func DBClearRecentFlag(ctx context.Context, tx *ent.Tx, mboxID imap.InternalMailboxID, messageID imap.InternalMessageID) error {
	if _, err := tx.UID.Update().
		Where(
			uid.HasMailboxWith(mailbox.MailboxID(mboxID)),
			uid.HasMessageWith(message.MessageID(messageID)),
			uid.Recent(true),
		).
		SetRecent(false).
		Save(ctx); err != nil {
		return err
	}

	return nil
}

func DBClearRecentFlags(ctx context.Context, tx *ent.Tx, mboxID imap.InternalMailboxID) error {
	if _, err := tx.UID.Update().
		Where(
			uid.HasMailboxWith(mailbox.MailboxID(mboxID)),
			uid.Recent(true),
		).
		SetRecent(false).
		Save(ctx); err != nil {
		return err
	}

	return nil
}

func DBUpdateRemoteMessageID(ctx context.Context, tx *ent.Tx, internalID imap.InternalMessageID, remoteID imap.MessageID) error {
	if _, err := tx.Message.Update().
		Where(message.MessageID(internalID)).
		SetRemoteID(remoteID).
		Save(ctx); err != nil {
		return err
	}

	return nil
}

func DBMarkMessageAsDeleted(ctx context.Context, tx *ent.Tx, messageID imap.InternalMessageID) error {
	if _, err := tx.Message.Update().Where(message.MessageID(messageID)).SetDeleted(true).Save(ctx); err != nil {
		return err
	}

	return nil
}

func DBMarkMessageAsDeletedWithRemoteID(ctx context.Context, tx *ent.Tx, messageID imap.MessageID) error {
	if _, err := tx.Message.Update().Where(message.RemoteID(messageID)).SetDeleted(true).Save(ctx); err != nil {
		return err
	}

	return nil
}

func DBDeleteMessages(ctx context.Context, tx *ent.Tx, messageIDs ...imap.InternalMessageID) error {
	if _, err := tx.Message.Delete().Where(message.MessageIDIn(messageIDs...)).Exec(ctx); err != nil {
		return err
	}

	return nil
}

func DBGetMessageIDsMarkedDeleted(ctx context.Context, client *ent.Client) ([]imap.InternalMessageID, error) {
	messages, err := client.Message.Query().Where(message.Deleted(true)).Select(message.FieldMessageID).All(ctx)
	if err != nil {
		return nil, err
	}

	return xslices.Map(messages, func(t *ent.Message) imap.InternalMessageID {
		return t.MessageID
	}), nil
}

func DBHasMessageWithID(ctx context.Context, client *ent.Client, id imap.InternalMessageID) (bool, error) {
	count, err := client.Message.Query().Where(message.MessageID(id)).Count(ctx)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func DBGetMessageIDFromRemoteID(ctx context.Context, client *ent.Client, id imap.MessageID) (imap.InternalMessageID, error) {
	message, err := client.Message.Query().Where(message.RemoteID(id)).Select(message.FieldMessageID).Only(ctx)
	if err != nil {
		return "", err
	}

	return message.MessageID, nil
}
