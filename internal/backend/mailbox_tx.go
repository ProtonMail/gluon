package backend

import (
	"context"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/backend/ent"
	"github.com/ProtonMail/gluon/internal/backend/ent/mailbox"
	"github.com/ProtonMail/gluon/internal/backend/ent/message"
	"github.com/ProtonMail/gluon/internal/backend/ent/uid"
	"github.com/bradenaw/juniper/xslices"
)

func txCreateMailbox(ctx context.Context, tx *ent.Tx, mboxID, name string, flags, permFlags, attrs imap.FlagSet) (*ent.Mailbox, error) {
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

	return create.Save(ctx)
}

func txMailboxExists(ctx context.Context, tx *ent.Tx, mboxID string) (bool, error) {
	count, err := tx.Mailbox.Query().Where(mailbox.MailboxID(mboxID)).Count(ctx)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func txRenameMailbox(ctx context.Context, tx *ent.Tx, mboxID, name string) error {
	if _, err := tx.Mailbox.Update().
		Where(mailbox.MailboxID(mboxID)).
		SetName(name).
		Save(ctx); err != nil {
		return err
	}

	return nil
}

func txDeleteMailbox(ctx context.Context, tx *ent.Tx, mboxID string) error {
	if _, err := tx.Mailbox.Delete().
		Where(mailbox.MailboxID(mboxID)).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func txUpdateMailboxID(ctx context.Context, tx *ent.Tx, oldID, newID string) error {
	if _, err := tx.Mailbox.Update().
		Where(mailbox.MailboxID(oldID)).
		SetMailboxID(newID).
		Save(ctx); err != nil {
		return err
	}

	return nil
}

func txBumpMailboxUIDNext(ctx context.Context, tx *ent.Tx, mbox *ent.Mailbox, withCount ...int) error {
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

func txGetMailboxName(ctx context.Context, tx *ent.Tx, mboxID string) (string, error) {
	mailbox, err := tx.Mailbox.Query().Where(mailbox.MailboxID(mboxID)).Select(mailbox.FieldName).Only(ctx)
	if err != nil {
		return "", err
	}

	return mailbox.Name, nil
}

func txGetMailboxMessageIDs(ctx context.Context, tx *ent.Tx, mailboxID string) ([]string, error) {
	messages, err := tx.Message.Query().
		Where(message.HasUIDsWith(uid.HasMailboxWith(mailbox.MailboxID(mailboxID)))).
		Select(message.FieldMessageID).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return xslices.Map(messages, func(message *ent.Message) string {
		return message.MessageID
	}), nil
}
