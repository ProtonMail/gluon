package db

import (
	"context"
	"fmt"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db/ent"
	"github.com/ProtonMail/gluon/internal/db/ent/mailbox"
	"github.com/ProtonMail/gluon/internal/db/ent/message"
	"github.com/ProtonMail/gluon/internal/db/ent/messageflag"
	"github.com/ProtonMail/gluon/internal/db/ent/uid"
	"github.com/ProtonMail/gluon/internal/ids"
	"github.com/bradenaw/juniper/xslices"
)

func CreateMailbox(ctx context.Context, tx *ent.Tx, labelID imap.LabelID, name string, flags, permFlags, attrs imap.FlagSet) (*ent.Mailbox, error) {
	create := tx.Mailbox.Create().
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

func RenameMailboxWithRemoteID(ctx context.Context, tx *ent.Tx, mboxID imap.LabelID, name string) error {
	if _, err := tx.Mailbox.Update().
		Where(mailbox.RemoteID(mboxID)).
		SetName(name).
		Save(ctx); err != nil {
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

	if _, err := tx.Mailbox.Update().Where(mailbox.ID(mbox.ID)).
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

func GetMailboxByRemoteID(ctx context.Context, client *ent.Client, id imap.LabelID) (*ent.Mailbox, error) {
	return client.Mailbox.Query().Where(mailbox.RemoteID(id)).Only(ctx)
}

func GetMailboxRecentCount(ctx context.Context, client *ent.Client, mbox *ent.Mailbox) (int, error) {
	return mbox.QueryUIDs().Where(uid.Recent(true)).Count(ctx)
}

type SnapshotMessageResult struct {
	InternalID imap.InternalMessageID `json:"uid_message"`
	RemoteID   imap.MessageID         `json:"remote_id"`
	UID        imap.UID               `json:"uid"`
	Recent     bool                   `json:"recent"`
	Deleted    bool                   `json:"deleted"`
	Flags      string                 `json:"flags"`
}

func (msg *SnapshotMessageResult) GetFlagSet() imap.FlagSet {
	var flagSet imap.FlagSet

	if len(msg.Flags) > 0 {
		flags := strings.Split(msg.Flags, ",")
		flagSet = imap.NewFlagSetFromSlice(flags)
	} else {
		flagSet = imap.NewFlagSet()
	}

	if msg.Deleted {
		flagSet = flagSet.Add(imap.FlagDeleted)
	}

	if msg.Recent {
		flagSet = flagSet.Add(imap.FlagRecent)
	}

	return flagSet
}

func GetMailboxMessagesForNewSnapshot(ctx context.Context, client *ent.Client, mboxID imap.InternalMailboxID) ([]SnapshotMessageResult, error) {
	messages := make([]SnapshotMessageResult, 0, 32)

	if err := client.UID.Query().Where(func(s *sql.Selector) {
		msgTable := sql.Table(message.Table)
		flagTable := sql.Table(messageflag.Table)
		s.Join(msgTable).On(s.C(uid.MessageColumn), msgTable.C(message.FieldID))
		s.LeftJoin(flagTable).On(s.C(uid.MessageColumn), flagTable.C(messageflag.MessagesColumn))
		s.Where(sql.EQ(uid.MailboxColumn, mboxID))
		s.Select(msgTable.C(message.FieldRemoteID), sql.As(fmt.Sprintf("GROUP_CONCAT(%v)", flagTable.C(messageflag.FieldValue)), "flags"), s.C(uid.FieldRecent), s.C(uid.FieldDeleted), s.C(uid.FieldUID), s.C(uid.MessageColumn))
		s.GroupBy(s.C(uid.MessageColumn))
		s.OrderBy(s.C(uid.FieldUID))
	}).Select().Scan(ctx, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

func GetMailboxIDWithRemoteID(ctx context.Context, client *ent.Client, labelID imap.LabelID) (imap.InternalMailboxID, error) {
	mbox, err := client.Mailbox.Query().Where(mailbox.RemoteID(labelID)).Select(mailbox.FieldID).Only(ctx)
	if err != nil {
		return 0, err
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

func CreateMailboxIfNotExists(ctx context.Context, tx *ent.Tx, mbox imap.Mailbox, delimiter string) error {
	exists, err := MailboxExistsWithRemoteID(ctx, tx.Client(), mbox.ID)
	if err != nil {
		return err
	}

	if !exists {
		if _, err := CreateMailbox(
			ctx,
			tx,
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
