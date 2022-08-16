package backend

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/backend/ent"
	"github.com/ProtonMail/gluon/internal/backend/ent/mailbox"
	"github.com/ProtonMail/gluon/internal/backend/ent/message"
	"github.com/ProtonMail/gluon/internal/backend/ent/uid"
	"github.com/bradenaw/juniper/xslices"
)

func DBCreateMailbox(ctx context.Context, tx *ent.Tx, mboxID imap.InternalMailboxID, labelID imap.LabelID, name string, flags, permFlags, attrs imap.FlagSet) (*ent.Mailbox, error) {
	create := tx.Mailbox.Create().
		SetMailboxID(mboxID).
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

func DBMailboxExistsWithID(ctx context.Context, client *ent.Client, mboxID imap.InternalMailboxID) (bool, error) {
	return client.Mailbox.Query().Where(mailbox.MailboxID(mboxID)).Exist(ctx)
}

func DBMailboxExistsWithRemoteID(ctx context.Context, client *ent.Client, mboxID imap.LabelID) (bool, error) {
	return client.Mailbox.Query().Where(mailbox.RemoteID(mboxID)).Exist(ctx)
}

func DBMailboxExistsWithName(ctx context.Context, client *ent.Client, name string) (bool, error) {
	return client.Mailbox.Query().Where(mailbox.Name(name)).Exist(ctx)
}

func DBRenameMailbox(ctx context.Context, tx *ent.Tx, mboxID imap.InternalMailboxID, name string) error {
	if _, err := tx.Mailbox.Update().
		Where(mailbox.MailboxID(mboxID)).
		SetName(name).
		Save(ctx); err != nil {
		return err
	}

	return nil
}

func DBRenameMailboxWithRemoteID(ctx context.Context, tx *ent.Tx, mboxID imap.LabelID, name string) error {
	if _, err := tx.Mailbox.Update().
		Where(mailbox.RemoteID(mboxID)).
		SetName(name).
		Save(ctx); err != nil {
		return err
	}

	return nil
}

func DBDeleteMailbox(ctx context.Context, tx *ent.Tx, mboxID imap.InternalMailboxID) error {
	if _, err := tx.Mailbox.Delete().
		Where(mailbox.MailboxID(mboxID)).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func DBDeleteMailboxWithRemoteID(ctx context.Context, tx *ent.Tx, mboxID imap.LabelID) error {
	if _, err := tx.Mailbox.Delete().
		Where(mailbox.RemoteID(mboxID)).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func DBUpdateRemoteMailboxID(ctx context.Context, tx *ent.Tx, internalID imap.InternalMailboxID, remoteID imap.LabelID) error {
	if _, err := tx.Mailbox.Update().
		Where(mailbox.MailboxID(internalID)).
		SetRemoteID(remoteID).
		Save(ctx); err != nil {
		return err
	}

	return nil
}

func DBBumpMailboxUIDNext(ctx context.Context, tx *ent.Tx, mbox *ent.Mailbox, withCount ...int) error {
	var n int

	if len(withCount) > 0 {
		n = withCount[0]
	} else {
		n = 1
	}

	if _, err := mbox.Update().
		SetUIDNext(mbox.UIDNext + n).
		Save(ctx); err != nil {
		return err
	}

	return nil
}

func DBGetMailboxName(ctx context.Context, client *ent.Client, mboxID imap.InternalMailboxID) (string, error) {
	mailbox, err := client.Mailbox.Query().Where(mailbox.MailboxID(mboxID)).Select(mailbox.FieldName).Only(ctx)
	if err != nil {
		return "", err
	}

	return mailbox.Name, nil
}

func DBGetMailboxNameWithRemoteID(ctx context.Context, client *ent.Client, mboxID imap.LabelID) (string, error) {
	mailbox, err := client.Mailbox.Query().Where(mailbox.RemoteID(mboxID)).Select(mailbox.FieldName).Only(ctx)
	if err != nil {
		return "", err
	}

	return mailbox.Name, nil
}

func DBGetMailboxMessageIDs(ctx context.Context, client *ent.Client, mailboxID imap.InternalMailboxID) ([]imap.InternalMessageID, error) {
	messages, err := client.Message.Query().
		Where(message.HasUIDsWith(uid.HasMailboxWith(mailbox.MailboxID(mailboxID)))).
		Select(message.FieldMessageID).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return xslices.Map(messages, func(message *ent.Message) imap.InternalMessageID {
		return message.MessageID
	}), nil
}

func DBGetMailboxMessageIDPairs(ctx context.Context, client *ent.Client, mailboxID imap.InternalMailboxID) ([]MessageIDPair, error) {
	messages, err := client.Message.Query().
		Where(message.HasUIDsWith(uid.HasMailboxWith(mailbox.MailboxID(mailboxID)))).
		Select(message.FieldMessageID, message.FieldRemoteID).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return xslices.Map(messages, func(message *ent.Message) MessageIDPair {
		return NewMessageIDPair(message)
	}), nil
}

func DBGetUIDInterval(ctx context.Context, mbox *ent.Mailbox, begin, end int) ([]*ent.UID, error) {
	return mbox.QueryUIDs().
		Where(uid.UIDGTE(begin), uid.UIDLTE(end)).
		WithMessage().
		All(ctx)
}

func DBGetAllMailboxes(ctx context.Context, client *ent.Client) ([]*ent.Mailbox, error) {
	const QueryLimit = 16000

	var mailboxes []*ent.Mailbox

	queryOffset := 0

	for i := 0; ; i += QueryLimit {
		result, err := client.Mailbox.Query().
			Where(mailbox.IDGT(queryOffset)).
			WithAttributes().
			Limit(QueryLimit).
			Order(func(selector *sql.Selector) { selector.OrderBy(mailbox.FieldID) }).
			All(ctx)
		if err != nil {
			return nil, err
		}

		resultLen := len(result)
		if resultLen == 0 {
			break
		}

		queryOffset = result[resultLen-1].ID
		mailboxes = append(mailboxes, result...)
	}

	return mailboxes, nil
}

func DBGetMailboxByName(ctx context.Context, client *ent.Client, name string) (*ent.Mailbox, error) {
	return client.Mailbox.Query().Where(mailbox.Name(name)).Only(ctx)
}

func DBGetMailboxByID(ctx context.Context, client *ent.Client, id imap.InternalMailboxID) (*ent.Mailbox, error) {
	return client.Mailbox.Query().Where(mailbox.MailboxID(id)).Only(ctx)
}

func DBGetMailboxMessages(ctx context.Context, mbox *ent.Mailbox) ([]*ent.UID, error) {
	return mbox.QueryUIDs().WithMessage().All(ctx)
}

func DBGetMailboxRecentCount(ctx context.Context, mbox *ent.Mailbox) (int, error) {
	return mbox.QueryUIDs().Where(uid.Recent(true)).Count(ctx)
}

func DBGetMailboxMessagesForNewSnapshot(ctx context.Context, mbox *ent.Mailbox) ([]*ent.UID, error) {
	var msgUIDs []*ent.UID

	const QueryLimit = 16000

	queryOffset := 0

	for i := 0; ; i += QueryLimit {
		result, err := mbox.QueryUIDs().
			Where(uid.IDGT(queryOffset)).
			WithMessage(func(query *ent.MessageQuery) { query.WithFlags().Select(message.FieldMessageID, message.FieldRemoteID) }).
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

func DBGetMailboxMessage(ctx context.Context, client *ent.Client, mailboxID imap.InternalMailboxID, messageID imap.InternalMessageID) (*ent.UID, error) {
	return client.UID.Query().
		Where(
			uid.HasMailboxWith(mailbox.MailboxID(mailboxID)),
			uid.HasMessageWith(message.MessageID(messageID)),
		).
		WithMessage(func(query *ent.MessageQuery) { query.WithFlags() }).
		Only(ctx)
}

func DBGetMailboxIDWithRemoteID(ctx context.Context, client *ent.Client, labelID imap.LabelID) (imap.InternalMailboxID, error) {
	mbox, err := client.Mailbox.Query().Where(mailbox.RemoteID(labelID)).Select(mailbox.FieldMailboxID).Only(ctx)
	if err != nil {
		return "", err
	}

	return mbox.MailboxID, nil
}

func DBTranslateRemoteMailboxIDs(ctx context.Context, client *ent.Client, mboxIDs []imap.LabelID) ([]imap.InternalMailboxID, error) {
	mboxes, err := client.Mailbox.Query().Where(mailbox.RemoteIDIn(mboxIDs...)).Select(mailbox.FieldMailboxID).All(ctx)
	if err != nil {
		return nil, err
	}

	return xslices.Map(mboxes, func(m *ent.Mailbox) imap.InternalMailboxID {
		return m.MailboxID
	}), nil
}
