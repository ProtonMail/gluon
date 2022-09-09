package db

import (
	"context"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db/ent"
	"github.com/ProtonMail/gluon/internal/db/ent/mailbox"
	"github.com/ProtonMail/gluon/internal/db/ent/message"
	"github.com/ProtonMail/gluon/internal/db/ent/uid"
	"github.com/ProtonMail/gluon/internal/ids"
	"github.com/bradenaw/juniper/xslices"
)

func CreateMailbox(ctx context.Context, tx *ent.Tx, mboxID imap.InternalMailboxID, labelID imap.LabelID, name string, flags, permFlags, attrs imap.FlagSet) (*ent.Mailbox, error) {
	create := tx.Mailbox.Create().
		SetID(mboxID).
		SetName(name)

	for _, flag := range flags.ToSlice() {
		create.AddFlags(tx.MailboxFlag.Create().SetValue(flag).SaveX(ctx))
	}

	for _, flag := range permFlags.ToSlice() {
		create.AddPermanentFlags(tx.MailboxPermFlag.Create().SetValue(flag).SaveX(ctx))
	}

	for _, attr := range attrs.ToSlice() {
		create.AddAttributes(tx.MailboxAttr.Create().SetValue(attr).SaveX(ctx))
	}

	if len(labelID) != 0 {
		create = create.SetRemoteID(labelID)
	}

	return create.Save(ctx)
}

func MailboxExistsWithID(ctx context.Context, client *ent.Client, mboxID imap.InternalMailboxID) (bool, error) {
	return client.Mailbox.Query().Where(mailbox.ID(mboxID)).Exist(ctx)
}

func MailboxExistsWithRemoteID(ctx context.Context, client *ent.Client, mboxID imap.LabelID) (bool, error) {
	return client.Mailbox.Query().Where(mailbox.RemoteID(mboxID)).Exist(ctx)
}

func MailboxExistsWithName(ctx context.Context, client *ent.Client, name string) (bool, error) {
	return client.Mailbox.Query().Where(mailbox.Name(name)).Exist(ctx)
}

func RenameMailbox(ctx context.Context, tx *ent.Tx, mboxID imap.InternalMailboxID, name string) error {
	if _, err := tx.Mailbox.Update().
		Where(mailbox.ID(mboxID)).
		SetName(name).
		Save(ctx); err != nil {
		return err
	}

	return nil
}

func RenameMailboxWithRemoteID(ctx context.Context, tx *ent.Tx, mboxID imap.LabelID, name string) error {
	if _, err := tx.Mailbox.Update().
		Where(mailbox.RemoteID(mboxID)).
		SetName(name).
		Save(ctx); err != nil {
		return err
	}

	return nil
}

func DeleteMailbox(ctx context.Context, tx *ent.Tx, mboxID imap.InternalMailboxID) error {
	if _, err := tx.Mailbox.Delete().
		Where(mailbox.ID(mboxID)).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func DeleteMailboxWithRemoteID(ctx context.Context, tx *ent.Tx, mboxID imap.LabelID) error {
	if _, err := tx.Mailbox.Delete().
		Where(mailbox.RemoteID(mboxID)).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func UpdateRemoteMailboxID(ctx context.Context, tx *ent.Tx, internalID imap.InternalMailboxID, remoteID imap.LabelID) error {
	if _, err := tx.Mailbox.Update().
		Where(mailbox.ID(internalID)).
		SetRemoteID(remoteID).
		Save(ctx); err != nil {
		return err
	}

	return nil
}

func BumpMailboxUIDNext(ctx context.Context, tx *ent.Tx, mbox *ent.Mailbox, withCount ...int) error {
	var n int

	if len(withCount) > 0 {
		n = withCount[0]
	} else {
		n = 1
	}

	if _, err := mbox.Update().
		SetUIDNext(mbox.UIDNext.Add(uint32(n))).
		Save(ctx); err != nil {
		return err
	}

	return nil
}

func GetMailboxName(ctx context.Context, client *ent.Client, mboxID imap.InternalMailboxID) (string, error) {
	mailbox, err := client.Mailbox.Query().Where(mailbox.ID(mboxID)).Select(mailbox.FieldName).Only(ctx)
	if err != nil {
		return "", err
	}

	return mailbox.Name, nil
}

func GetMailboxNameWithRemoteID(ctx context.Context, client *ent.Client, mboxID imap.LabelID) (string, error) {
	mailbox, err := client.Mailbox.Query().Where(mailbox.RemoteID(mboxID)).Select(mailbox.FieldName).Only(ctx)
	if err != nil {
		return "", err
	}

	return mailbox.Name, nil
}

func GetMailboxMessageIDs(ctx context.Context, client *ent.Client, mailboxID imap.InternalMailboxID) ([]imap.InternalMessageID, error) {
	messages, err := client.Message.Query().
		Where(message.HasUIDsWith(uid.HasMailboxWith(mailbox.ID(mailboxID)))).
		Select(message.FieldID).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return xslices.Map(messages, func(message *ent.Message) imap.InternalMessageID {
		return message.ID
	}), nil
}

func GetMailboxMessageIDPairs(ctx context.Context, client *ent.Client, mailboxID imap.InternalMailboxID) ([]ids.MessageIDPair, error) {
	messages, err := client.Message.Query().
		Where(message.HasUIDsWith(uid.HasMailboxWith(mailbox.ID(mailboxID)))).
		Select(message.FieldID, message.FieldRemoteID).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return xslices.Map(messages, func(message *ent.Message) ids.MessageIDPair {
		return ids.NewMessageIDPair(message)
	}), nil
}

func GetUIDInterval(ctx context.Context, client *ent.Client, mbox *ent.Mailbox, begin, end imap.UID) ([]*ent.UID, error) {
	return mbox.QueryUIDs().
		Where(uid.UIDGTE(begin), uid.UIDLTE(end)).
		WithMessage().
		All(ctx)
}

func GetAllMailboxes(ctx context.Context, client *ent.Client) ([]*ent.Mailbox, error) {
	const QueryLimit = 16000

	var mailboxes []*ent.Mailbox

	for i := 0; ; i += QueryLimit {
		result, err := client.Mailbox.Query().
			WithAttributes().
			Limit(QueryLimit).
			Offset(i).
			All(ctx)
		if err != nil {
			return nil, err
		}

		resultLen := len(result)
		if resultLen == 0 {
			break
		}

		mailboxes = append(mailboxes, result...)
	}

	return mailboxes, nil
}

func GetMailboxByName(ctx context.Context, client *ent.Client, name string) (*ent.Mailbox, error) {
	return client.Mailbox.Query().Where(mailbox.Name(name)).Only(ctx)
}

func GetMailboxByID(ctx context.Context, client *ent.Client, id imap.InternalMailboxID) (*ent.Mailbox, error) {
	return client.Mailbox.Query().Where(mailbox.ID(id)).Only(ctx)
}

func GetMailboxMessages(ctx context.Context, client *ent.Client, mbox *ent.Mailbox) ([]*ent.UID, error) {
	return mbox.QueryUIDs().WithMessage().All(ctx)
}

func GetMailboxRecentCount(ctx context.Context, client *ent.Client, mbox *ent.Mailbox) (int, error) {
	return mbox.QueryUIDs().Where(uid.Recent(true)).Count(ctx)
}

func GetMailboxMessagesForNewSnapshot(ctx context.Context, client *ent.Client, mbox *ent.Mailbox) ([]*ent.UID, error) {
	var msgUIDs []*ent.UID

	const QueryLimit = 16000

	queryOffset := 0

	for i := 0; ; i += QueryLimit {
		result, err := mbox.QueryUIDs().
			Where(uid.IDGT(queryOffset)).
			WithMessage(func(query *ent.MessageQuery) { query.WithFlags().Select(message.FieldID, message.FieldRemoteID) }).
			Select(uid.FieldID, uid.FieldUID, uid.FieldRecent, uid.FieldDeleted).Order(func(selector *sql.Selector) {
			selector.OrderBy(uid.FieldID)
		}).Limit(QueryLimit).All(ctx)
		if err != nil {
			return nil, err
		}

		resultLen := len(result)

		if resultLen == 0 {
			break
		}

		queryOffset = result[resultLen-1].ID
		msgUIDs = append(msgUIDs, result...)
	}

	return msgUIDs, nil
}

func GetMailboxMessage(ctx context.Context, client *ent.Client, mailboxID imap.InternalMailboxID, messageID imap.InternalMessageID) (*ent.UID, error) {
	return client.UID.Query().
		Where(
			uid.HasMailboxWith(mailbox.ID(mailboxID)),
			uid.HasMessageWith(message.ID(messageID)),
		).
		WithMessage(func(query *ent.MessageQuery) { query.WithFlags() }).
		Only(ctx)
}

func GetMailboxIDWithRemoteID(ctx context.Context, client *ent.Client, labelID imap.LabelID) (imap.InternalMailboxID, error) {
	mbox, err := client.Mailbox.Query().Where(mailbox.RemoteID(labelID)).Select(mailbox.FieldID).Only(ctx)
	if err != nil {
		return "", err
	}

	return mbox.ID, nil
}

func TranslateRemoteMailboxIDs(ctx context.Context, client *ent.Client, mboxIDs []imap.LabelID) ([]imap.InternalMailboxID, error) {
	mboxes, err := client.Mailbox.Query().Where(mailbox.RemoteIDIn(mboxIDs...)).Select(mailbox.FieldID).All(ctx)
	if err != nil {
		return nil, err
	}

	return xslices.Map(mboxes, func(m *ent.Mailbox) imap.InternalMailboxID {
		return m.ID
	}), nil
}

func CreateMailboxIfNotExists(ctx context.Context, tx *ent.Tx, internalID imap.InternalMailboxID, mbox imap.Mailbox, delimiter string) error {
	exists, err := MailboxExistsWithID(ctx, tx.Client(), internalID)
	if err != nil {
		return err
	}

	if !exists {
		if _, err := CreateMailbox(
			ctx,
			tx,
			internalID,
			mbox.ID,
			strings.Join(mbox.Name, delimiter),
			mbox.Flags,
			mbox.PermanentFlags,
			mbox.Attributes,
		); err != nil {
			return err
		}
	}

	return nil
}
