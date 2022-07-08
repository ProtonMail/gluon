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
	literal    []byte
	body       string
	structure  string
	envelope   string
	internalID string
}

func DBCreateMessages(ctx context.Context, tx *ent.Tx, reqs ...*DBCreateMessageReq) ([]*ent.Message, error) {
	flags := make(map[string][]*ent.MessageFlag)

	for _, req := range reqs {
		builders := xslices.Map(req.message.Flags.ToSlice(), func(flag string) *ent.MessageFlagCreate {
			return tx.MessageFlag.Create().SetValue(flag)
		})

		entFlags, err := tx.MessageFlag.CreateBulk(builders...).Save(ctx)
		if err != nil {
			return nil, err
		}

		flags[req.message.ID] = entFlags
	}

	return tx.Message.CreateBulk(xslices.Map(reqs, func(req *DBCreateMessageReq) *ent.MessageCreate {
		return tx.Message.Create().
			SetMessageID(req.message.ID).
			SetInternalID(req.internalID).
			SetDate(req.message.Date).
			SetBody(req.body).
			SetBodyStructure(req.structure).
			SetEnvelope(req.envelope).
			SetSize(len(req.literal)).
			AddFlags(flags[req.message.ID]...)
	})...).Save(ctx)
}

func DBAddMessagesToMailbox(ctx context.Context, tx *ent.Tx, messageIDs []string, mboxID string) (map[string]int, error) {
	messageUIDs := make(map[string]int)

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

func DBRemoveMessagesFromMailbox(ctx context.Context, tx *ent.Tx, messageIDs []string, mboxID string) error {
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

func DBMessageExists(ctx context.Context, client *ent.Client, messageID string) (bool, error) {
	count, err := client.Message.Query().Where(message.MessageID(messageID)).Count(ctx)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func DBGetMessage(ctx context.Context, client *ent.Client, messageID string) (*ent.Message, error) {
	return client.Message.Query().Where(message.MessageID(messageID)).Only(ctx)
}

func DBGetMessages(ctx context.Context, client *ent.Client, messageIDs ...string) (map[string]*ent.Message, error) {
	messages, err := client.Message.Query().Where(message.MessageIDIn(messageIDs...)).All(ctx)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*ent.Message)

	for _, message := range messages {
		res[message.MessageID] = message
	}

	return res, nil
}

func DBGetMessageID(ctx context.Context, client *ent.Client, internalID string) (string, error) {
	message, err := client.Message.Query().Where(message.InternalID(internalID)).Select(message.FieldID).Only(ctx)
	if err != nil {
		return "", err
	}

	return message.MessageID, nil
}

func DBGetMessageMailboxIDs(ctx context.Context, client *ent.Client, messageID string) ([]string, error) {
	messageUIDs, err := client.UID.Query().
		Where(uid.HasMessageWith(message.MessageID(messageID))).
		WithMailbox(func(query *ent.MailboxQuery) {
			query.Select(mailbox.FieldMailboxID)
		}).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return xslices.Map(messageUIDs, func(message *ent.UID) string {
		return message.Edges.Mailbox.MailboxID
	}), nil
}

func DBGetMessageUIDs(ctx context.Context, client *ent.Client, mboxID string, messageIDs []string) (map[string]int, error) {
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

	res := make(map[string]int)

	for _, messageUID := range messageUIDs {
		res[messageUID.Edges.Message.MessageID] = messageUID.UID
	}

	return res, nil
}

// DBGetMessageFlags returns the flags of the given messages.
// It does not include per-mailbox flags (\Deleted, \Recent)!
func DBGetMessageFlags(ctx context.Context, client *ent.Client, messageIDs []string) (map[string]imap.FlagSet, error) {
	messages, err := client.Message.Query().
		Where(message.MessageIDIn(messageIDs...)).
		WithFlags(func(query *ent.MessageFlagQuery) {
			query.Select(messageflag.FieldValue)
		}).
		All(ctx)
	if err != nil {
		return nil, err
	}

	curFlags := make(map[string]imap.FlagSet)

	for _, message := range messages {
		curFlags[message.MessageID] = imap.NewFlagSetFromSlice(xslices.Map(message.Edges.Flags, func(flag *ent.MessageFlag) string {
			return flag.Value
		}))
	}

	return curFlags, nil
}

func DBGetMessageDeleted(ctx context.Context, client *ent.Client, mboxID string, messageIDs []string) (map[string]bool, error) {
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

	res := make(map[string]bool)

	for _, messageUID := range messageUIDs {
		res[messageUID.Edges.Message.MessageID] = messageUID.Deleted
	}

	return res, nil
}

func DBAddMessageFlag(ctx context.Context, tx *ent.Tx, messageIDs []string, addFlag string) error {
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

func DBRemoveMessageFlag(ctx context.Context, tx *ent.Tx, messageIDs []string, remFlag string) error {
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

func DBSetMessageFlags(ctx context.Context, tx *ent.Tx, messageIDs []string, setFlags imap.FlagSet) error {
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

func DBSetDeletedFlag(ctx context.Context, tx *ent.Tx, mboxID string, messageIDs []string, deleted bool) error {
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

func DBClearRecentFlag(ctx context.Context, tx *ent.Tx, mboxID, messageID string) error {
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

func DBClearRecentFlags(ctx context.Context, tx *ent.Tx, mboxID string) error {
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

func DBUpdateMessageID(ctx context.Context, tx *ent.Tx, oldID, newID string) error {
	if _, err := tx.Message.Update().
		Where(message.MessageID(oldID)).
		SetMessageID(newID).
		Save(ctx); err != nil {
		return err
	}

	return nil
}

func DBMarkMessageAsDeleted(ctx context.Context, tx *ent.Tx, messageID string) error {
	if _, err := tx.Message.Update().Where(message.MessageID(messageID)).SetDeleted(true).Save(ctx); err != nil {
		return err
	}

	return nil
}

func DBDeleteMessages(ctx context.Context, tx *ent.Tx, messageIDs ...string) error {
	if _, err := tx.Message.Delete().Where(message.MessageIDIn(messageIDs...)).Exec(ctx); err != nil {
		return err
	}

	return nil
}

func DBGetMessageIDsMarkedDeleted(ctx context.Context, tx *ent.Tx) ([]string, error) {
	messages, err := tx.Message.Query().Where(message.Deleted(true)).Select(message.FieldMessageID).All(ctx)
	if err != nil {
		return nil, err
	}

	return xslices.Map(messages, func(t *ent.Message) string {
		return t.MessageID
	}), nil
}
