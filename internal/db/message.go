package db

import (
	"context"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db/ent"
	"github.com/ProtonMail/gluon/internal/db/ent/mailbox"
	"github.com/ProtonMail/gluon/internal/db/ent/message"
	"github.com/ProtonMail/gluon/internal/db/ent/messageflag"
	"github.com/ProtonMail/gluon/internal/db/ent/uid"
	"github.com/bradenaw/juniper/xslices"
)

type CreateMessageReq struct {
	Message    imap.Message
	InternalID imap.InternalMessageID
	Literal    []byte
	Body       string
	Structure  string
	Envelope   string
}

func CreateMessages(ctx context.Context, tx *ent.Tx, reqs ...*CreateMessageReq) ([]*ent.Message, error) {
	flags := make(map[imap.InternalMessageID][]*ent.MessageFlag)

	for _, req := range reqs {
		builders := xslices.Map(req.Message.Flags.ToSlice(), func(flag string) *ent.MessageFlagCreate {
			return tx.MessageFlag.Create().SetValue(flag)
		})

		entFlags, err := tx.MessageFlag.CreateBulk(builders...).Save(ctx)
		if err != nil {
			return nil, err
		}

		flags[req.InternalID] = entFlags
	}

	return tx.Message.CreateBulk(xslices.Map(reqs, func(req *CreateMessageReq) *ent.MessageCreate {
		msgCreate := tx.Message.Create().
			SetMessageID(req.InternalID).
			SetDate(req.Message.Date).
			SetBody(req.Body).
			SetBodyStructure(req.Structure).
			SetEnvelope(req.Envelope).
			SetSize(len(req.Literal)).
			AddFlags(flags[req.InternalID]...)

		if len(req.Message.ID) != 0 {
			msgCreate = msgCreate.SetRemoteID(req.Message.ID)
		}

		return msgCreate
	})...).Save(ctx)
}

func AddMessagesToMailbox(ctx context.Context, tx *ent.Tx, messageIDs []imap.InternalMessageID, mboxID imap.InternalMailboxID) (map[imap.InternalMessageID]*ent.UID, error) {
	messageUIDs := make(map[imap.InternalMessageID]imap.UID)

	mbox, err := tx.Mailbox.Query().Where(mailbox.MailboxID(mboxID)).Only(ctx)
	if err != nil {
		return nil, err
	}

	messages, err := GetMessages(ctx, tx.Client(), messageIDs...)
	if err != nil {
		return nil, err
	}

	var builders []*ent.UIDCreate

	for idx, messageID := range messageIDs {
		messageUIDs[messageID] = mbox.UIDNext.Add(uint32(idx))

		builders = append(builders, tx.UID.Create().
			SetMailbox(mbox).
			SetMessage(messages[messageID]).
			SetUID(messageUIDs[messageID]),
		)
	}

	if _, err := tx.UID.CreateBulk(builders...).Save(ctx); err != nil {
		return nil, err
	}

	if err := BumpMailboxUIDNext(ctx, tx, mbox, len(messageIDs)); err != nil {
		return nil, err
	}

	return GetMessageUIDsWithFlags(ctx, tx.Client(), mboxID, messageIDs)
}

func BumpMailboxUIDsForMessage(ctx context.Context, tx *ent.Tx, messageIDs []imap.InternalMessageID, mboxID imap.InternalMailboxID) (map[imap.InternalMessageID]*ent.UID, error) {
	messageUIDs := make(map[imap.InternalMessageID]imap.UID)

	mbox, err := tx.Mailbox.Query().Where(mailbox.MailboxID(mboxID)).Only(ctx)
	if err != nil {
		return nil, err
	}

	var builders []*ent.UIDUpdate

	for idx, messageID := range messageIDs {
		uidNext := mbox.UIDNext.Add(uint32(idx))
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

	if err := BumpMailboxUIDNext(ctx, tx, mbox, len(messageIDs)); err != nil {
		return nil, err
	}

	return GetMessageUIDsWithFlags(ctx, tx.Client(), mboxID, messageIDs)
}

func RemoveMessagesFromMailbox(ctx context.Context, tx *ent.Tx, messageIDs []imap.InternalMessageID, mboxID imap.InternalMailboxID) error {
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

func MessageExists(ctx context.Context, client *ent.Client, messageID imap.InternalMessageID) (bool, error) {
	count, err := client.Message.Query().Where(message.MessageID(messageID)).Count(ctx)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func MessageExistsWithRemoteID(ctx context.Context, client *ent.Client, messageID imap.MessageID) (bool, error) {
	count, err := client.Message.Query().Where(message.RemoteID(messageID)).Count(ctx)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func GetMessage(ctx context.Context, client *ent.Client, messageID imap.InternalMessageID) (*ent.Message, error) {
	return client.Message.Query().Where(message.MessageID(messageID)).Only(ctx)
}

func GetMessages(ctx context.Context, client *ent.Client, messageIDs ...imap.InternalMessageID) (map[imap.InternalMessageID]*ent.Message, error) {
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

func GetRemoteMessageID(ctx context.Context, client *ent.Client, internalID imap.InternalMessageID) (imap.MessageID, error) {
	message, err := client.Message.Query().Where(message.MessageID(internalID)).Select(message.FieldRemoteID).Only(ctx)
	if err != nil {
		return "", err
	}

	return message.RemoteID, nil
}

func GetMessageMailboxIDs(ctx context.Context, client *ent.Client, messageID imap.InternalMessageID) ([]imap.InternalMailboxID, error) {
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

func GetMessageUIDs(ctx context.Context, client *ent.Client, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) (map[imap.InternalMessageID]imap.UID, error) {
	messageUIDs, err := client.UID.Query().
		Where(
			uid.HasMailboxWith(mailbox.MailboxID(mboxID)),
			uid.HasMessageWith(message.MessageIDIn(messageIDs...)),
		).
		WithMessage(func(query *ent.MessageQuery) {
			query.Select(message.FieldMessageID)
			query.Select(message.FieldRemoteID)
		}).
		All(ctx)
	if err != nil {
		return nil, err
	}

	res := make(map[imap.InternalMessageID]imap.UID)

	for _, messageUID := range messageUIDs {
		res[messageUID.Edges.Message.MessageID] = messageUID.UID
	}

	return res, nil
}

func GetMessageUIDsWithFlags(ctx context.Context, client *ent.Client, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) (map[imap.InternalMessageID]*ent.UID, error) {
	messageUIDs, err := client.UID.Query().
		Where(
			uid.HasMailboxWith(mailbox.MailboxID(mboxID)),
			uid.HasMessageWith(message.MessageIDIn(messageIDs...)),
		).
		WithMessage(func(query *ent.MessageQuery) {
			query.Select(message.FieldMessageID, message.FieldRemoteID)
			query.WithFlags(func(query *ent.MessageFlagQuery) {
				query.Select(messageflag.FieldValue)
			})
		}).
		All(ctx)
	if err != nil {
		return nil, err
	}

	res := make(map[imap.InternalMessageID]*ent.UID)

	for _, messageUID := range messageUIDs {
		res[messageUID.Edges.Message.MessageID] = messageUID
	}

	return res, nil
}

// GetMessageFlags returns the flags of the given messages.
// It does not include per-mailbox flags (\Deleted, \Recent)!
func GetMessageFlags(ctx context.Context, client *ent.Client, messageIDs []imap.InternalMessageID) (map[imap.InternalMessageID]imap.FlagSet, error) {
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

func GetMessageDeleted(ctx context.Context, client *ent.Client, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) (map[imap.InternalMessageID]bool, error) {
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

func AddMessageFlag(ctx context.Context, tx *ent.Tx, messageIDs []imap.InternalMessageID, addFlag string) error {
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

func RemoveMessageFlag(ctx context.Context, tx *ent.Tx, messageIDs []imap.InternalMessageID, remFlag string) error {
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

func SetMessageFlags(ctx context.Context, tx *ent.Tx, messageIDs []imap.InternalMessageID, setFlags imap.FlagSet) error {
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

func SetDeletedFlag(ctx context.Context, tx *ent.Tx, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID, deleted bool) error {
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

func ClearRecentFlag(ctx context.Context, tx *ent.Tx, mboxID imap.InternalMailboxID, messageID imap.InternalMessageID) error {
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

func ClearRecentFlags(ctx context.Context, tx *ent.Tx, mboxID imap.InternalMailboxID) error {
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

func UpdateRemoteMessageID(ctx context.Context, tx *ent.Tx, internalID imap.InternalMessageID, remoteID imap.MessageID) error {
	if _, err := tx.Message.Update().
		Where(message.MessageID(internalID)).
		SetRemoteID(remoteID).
		Save(ctx); err != nil {
		return err
	}

	return nil
}

func MarkMessageAsDeleted(ctx context.Context, tx *ent.Tx, messageID imap.InternalMessageID) error {
	if _, err := tx.Message.Update().Where(message.MessageID(messageID)).SetDeleted(true).Save(ctx); err != nil {
		return err
	}

	return nil
}

func MarkMessageAsDeletedWithRemoteID(ctx context.Context, tx *ent.Tx, messageID imap.MessageID) error {
	if _, err := tx.Message.Update().Where(message.RemoteID(messageID)).SetDeleted(true).Save(ctx); err != nil {
		return err
	}

	return nil
}

func DeleteMessages(ctx context.Context, tx *ent.Tx, messageIDs ...imap.InternalMessageID) error {
	if _, err := tx.Message.Delete().Where(message.MessageIDIn(messageIDs...)).Exec(ctx); err != nil {
		return err
	}

	return nil
}

func GetMessageIDsMarkedDeleted(ctx context.Context, client *ent.Client) ([]imap.InternalMessageID, error) {
	messages, err := client.Message.Query().Where(message.Deleted(true)).Select(message.FieldMessageID).All(ctx)
	if err != nil {
		return nil, err
	}

	return xslices.Map(messages, func(t *ent.Message) imap.InternalMessageID {
		return t.MessageID
	}), nil
}

func HasMessageWithID(ctx context.Context, client *ent.Client, id imap.InternalMessageID) (bool, error) {
	count, err := client.Message.Query().Where(message.MessageID(id)).Count(ctx)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func GetMessageIDFromRemoteID(ctx context.Context, client *ent.Client, id imap.MessageID) (imap.InternalMessageID, error) {
	message, err := client.Message.Query().Where(message.RemoteID(id)).Select(message.FieldMessageID).Only(ctx)
	if err != nil {
		return "", err
	}

	return message.MessageID, nil
}

func NewFlagSet(msgUID *ent.UID, flags []*ent.MessageFlag) imap.FlagSet {
	flagSet := imap.NewFlagSetFromSlice(xslices.Map(flags, func(flag *ent.MessageFlag) string {
		return flag.Value
	}))

	if msgUID.Deleted {
		flagSet = flagSet.Add(imap.FlagDeleted)
	}

	if msgUID.Recent {
		flagSet = flagSet.Add(imap.FlagRecent)
	}

	return flagSet
}
