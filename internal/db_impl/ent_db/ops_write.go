package ent_db

import (
	"context"
	"github.com/ProtonMail/gluon/db"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db_impl/ent_db/internal"
	"github.com/ProtonMail/gluon/internal/db_impl/ent_db/internal/mailbox"
	"github.com/bradenaw/juniper/xslices"
)

type EntOpsWrite struct {
	EntOpsRead
	tx *internal.Tx
}

func newEntOpsWrite(tx *internal.Tx) *EntOpsWrite {
	return &EntOpsWrite{
		EntOpsRead: EntOpsRead{client: tx.Client()},
		tx:         tx,
	}
}

func (op *EntOpsWrite) CreateMailbox(ctx context.Context,
	mboxID imap.MailboxID,
	name string,
	flags, permFlags, attrs imap.FlagSet,
	uidValidity imap.UID) (*db.Mailbox, error) {
	return wrapEntErrFnTyped(func() (*db.Mailbox, error) {
		mbox, err := CreateMailbox(ctx, op.tx, mboxID, name, flags, permFlags, attrs, uidValidity)

		return entMBoxToDB(mbox), err
	})
}

func (op *EntOpsWrite) GetOrCreateMailbox(ctx context.Context,
	mboxID imap.MailboxID,
	name string,
	flags, permFlags, attrs imap.FlagSet,
	uidValidity imap.UID) (*db.Mailbox, error) {
	return wrapEntErrFnTyped(func() (*db.Mailbox, error) {
		mbox, err := CreateMailbox(ctx, op.tx, mboxID, name, flags, permFlags, attrs, uidValidity)

		return entMBoxToDB(mbox), err
	})
}

func (op *EntOpsWrite) GetOrCreateMailboxAlt(ctx context.Context,
	mbox imap.Mailbox,
	delimiter string,
	uidValidity imap.UID) (*db.Mailbox, error) {
	return wrapEntErrFnTyped(func() (*db.Mailbox, error) {
		mbox, err := GetOrCreateMailbox(ctx, op.tx, mbox, delimiter, uidValidity)

		return entMBoxToDB(mbox), err
	})
}

func (op *EntOpsWrite) RenameMailboxWithRemoteID(ctx context.Context, mboxID imap.MailboxID, name string) error {
	return wrapEntErrFn(func() error {
		return RenameMailboxWithRemoteID(ctx, op.tx, mboxID, name)
	})
}

func (op *EntOpsWrite) DeleteMailboxWithRemoteID(ctx context.Context, mboxID imap.MailboxID) error {
	return wrapEntErrFn(func() error {
		return DeleteMailboxWithRemoteID(ctx, op.tx, mboxID)
	})
}

func (op *EntOpsWrite) BumpMailboxUIDNext(ctx context.Context, mboxID imap.InternalMailboxID, count int) error {
	return wrapEntErrFn(func() error {
		mbox, err := op.tx.Mailbox.Query().Where(mailbox.ID(mboxID)).Only(ctx)
		if err != nil {
			return err
		}

		return BumpMailboxUIDNext(ctx, op.tx, mbox, count)
	})
}

func (op *EntOpsWrite) AddMessagesToMailbox(ctx context.Context, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) ([]db.UIDWithFlags, error) {
	return wrapEntErrFnTyped(func() ([]db.UIDWithFlags, error) {
		return AddMessagesToMailbox(ctx, op.tx, messageIDs, mboxID)
	})
}

func (op *EntOpsWrite) BumpMailboxUIDsForMessage(ctx context.Context, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) ([]db.UIDWithFlags, error) {
	return wrapEntErrFnTyped(func() ([]db.UIDWithFlags, error) {
		return BumpMailboxUIDsForMessage(ctx, op.tx, messageIDs, mboxID)
	})
}

func (op *EntOpsWrite) RemoveMessagesFromMailbox(ctx context.Context, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) error {
	return wrapEntErrFn(func() error {
		return RemoveMessagesFromMailbox(ctx, op.tx, messageIDs, mboxID)
	})
}

func (op *EntOpsWrite) ClearRecentFlagInMailboxOnMessage(ctx context.Context, mboxID imap.InternalMailboxID, messageID imap.InternalMessageID) error {
	return wrapEntErrFn(func() error {
		return ClearRecentFlag(ctx, op.tx, mboxID, messageID)
	})
}

func (op *EntOpsWrite) ClearRecentFlagsInMailbox(ctx context.Context, mboxID imap.InternalMailboxID) error {
	return wrapEntErrFn(func() error {
		return ClearRecentFlags(ctx, op.tx, mboxID)
	})
}

func (op *EntOpsWrite) CreateMailboxIfNotExists(ctx context.Context, mbox imap.Mailbox, delimiter string, uidValidity imap.UID) error {
	return wrapEntErrFn(func() error {
		return CreateMailboxIfNotExists(ctx, op.tx, mbox, delimiter, uidValidity)
	})
}

func (op *EntOpsWrite) SetMailboxMessagesDeletedFlag(ctx context.Context, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID, deleted bool) error {
	return wrapEntErrFn(func() error {
		return SetDeletedFlag(ctx, op.tx, mboxID, messageIDs, deleted)
	})
}

func (op *EntOpsWrite) SetMailboxSubscribed(ctx context.Context, mboxID imap.InternalMailboxID, subscribed bool) error {
	return wrapEntErrFn(func() error {
		return op.tx.Mailbox.Update().Where(mailbox.ID(mboxID)).SetSubscribed(subscribed).Exec(ctx)
	})
}

func (op *EntOpsWrite) UpdateRemoteMailboxID(ctx context.Context, mobxID imap.InternalMailboxID, remoteID imap.MailboxID) error {
	return wrapEntErrFn(func() error {
		return UpdateRemoteMailboxID(ctx, op.tx, mobxID, remoteID)
	})
}

func (op *EntOpsWrite) SetMailboxUIDValidity(ctx context.Context, mboxID imap.InternalMailboxID, uidValidity imap.UID) error {
	return wrapEntErrFn(func() error {
		return op.tx.Mailbox.Update().Where(mailbox.ID(mboxID)).SetUIDValidity(uidValidity).Exec(ctx)
	})
}

func (op *EntOpsWrite) CreateMessages(ctx context.Context, reqs ...*db.CreateMessageReq) ([]*db.Message, error) {
	return wrapEntErrFnTyped(func() ([]*db.Message, error) {
		msgs, err := CreateMessages(ctx, op.tx, reqs...)

		return xslices.Map(msgs, entMessageToDB), err
	})
}

func (op *EntOpsWrite) CreateMessageAndAddToMailbox(ctx context.Context, mbox imap.InternalMailboxID, req *db.CreateMessageReq) (imap.UID, imap.FlagSet, error) {
	var uid imap.UID

	var flagSet imap.FlagSet

	err := wrapEntErrFn(func() error {
		var err error

		uid, flagSet, err = CreateAndAddMessageToMailbox(ctx, op.tx, mbox, req)

		return err
	})

	return uid, flagSet, err
}

func (op *EntOpsWrite) MarkMessageAsDeleted(ctx context.Context, id imap.InternalMessageID) error {
	return wrapEntErrFn(func() error {
		return MarkMessageAsDeleted(ctx, op.tx, id)
	})
}

func (op *EntOpsWrite) MarkMessageAsDeletedAndAssignRandomRemoteID(ctx context.Context, id imap.InternalMessageID) error {
	return wrapEntErrFn(func() error {
		return MarkMessageAsDeletedAndAssignRandomRemoteID(ctx, op.tx, id)
	})
}

func (op *EntOpsWrite) MarkMessageAsDeletedWithRemoteID(ctx context.Context, id imap.MessageID) error {
	return wrapEntErrFn(func() error {
		return MarkMessageAsDeletedWithRemoteID(ctx, op.tx, id)
	})
}

func (op *EntOpsWrite) DeleteMessages(ctx context.Context, ids []imap.InternalMessageID) error {
	return wrapEntErrFn(func() error {
		return DeleteMessages(ctx, op.tx, ids...)
	})
}

func (op *EntOpsWrite) UpdateRemoteMessageID(ctx context.Context, internalID imap.InternalMessageID, remoteID imap.MessageID) error {
	return wrapEntErrFn(func() error {
		return UpdateRemoteMessageID(ctx, op.tx, internalID, remoteID)
	})
}

func (op *EntOpsWrite) AddFlagToMessages(ctx context.Context, ids []imap.InternalMessageID, flag string) error {
	return wrapEntErrFn(func() error {
		return AddMessageFlag(ctx, op.tx, ids, flag)
	})
}

func (op *EntOpsWrite) RemoveFlagFromMessages(ctx context.Context, ids []imap.InternalMessageID, flag string) error {
	return wrapEntErrFn(func() error {
		return RemoveMessageFlag(ctx, op.tx, ids, flag)
	})
}

func (op *EntOpsWrite) SetFlagsOnMessages(ctx context.Context, ids []imap.InternalMessageID, flags imap.FlagSet) error {
	return wrapEntErrFn(func() error {
		return SetMessageFlags(ctx, op.tx, ids, flags)
	})
}

func (op *EntOpsWrite) AddDeletedSubscription(ctx context.Context, mboxName string, mboxID imap.MailboxID) error {
	return wrapEntErrFn(func() error {
		return AddDeletedSubscription(ctx, op.tx, mboxName, mboxID)
	})
}

func (op *EntOpsWrite) RemoveDeletedSubscriptionWithName(ctx context.Context, mboxName string) (int, error) {
	return wrapEntErrFnTyped(func() (int, error) {
		return RemoveDeletedSubscriptionWithName(ctx, op.tx, mboxName)
	})
}
