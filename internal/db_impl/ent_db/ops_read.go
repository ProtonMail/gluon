package ent_db

import (
	"context"
	"time"

	"github.com/ProtonMail/gluon/db"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db_impl/ent_db/internal"
	"github.com/bradenaw/juniper/xslices"
)

type EntOpsRead struct {
	client *internal.Client
}

func newOpsReadFromClient(client *internal.Client) *EntOpsRead {
	return &EntOpsRead{
		client: client,
	}
}

func newOpsReadFromTx(tx *internal.Tx) *EntOpsRead {
	return &EntOpsRead{
		client: tx.Client(),
	}
}

func (op *EntOpsRead) MailboxExistsWithID(ctx context.Context, mboxID imap.InternalMailboxID) (bool, error) {
	return wrapEntErrFnTyped(func() (bool, error) {
		return MailboxExistsWithID(ctx, op.client, mboxID)
	})
}

func (op *EntOpsRead) MailboxExistsWithRemoteID(ctx context.Context, mboxID imap.MailboxID) (bool, error) {
	return wrapEntErrFnTyped(func() (bool, error) {
		return MailboxExistsWithRemoteID(ctx, op.client, mboxID)
	})
}

func (op *EntOpsRead) MailboxExistsWithName(ctx context.Context, name string) (bool, error) {
	return wrapEntErrFnTyped(func() (bool, error) {
		return MailboxExistsWithName(ctx, op.client, name)
	})
}

func (op *EntOpsRead) GetMailboxIDFromRemoteID(ctx context.Context, mboxID imap.MailboxID) (imap.InternalMailboxID, error) {
	return wrapEntErrFnTyped(func() (imap.InternalMailboxID, error) {
		return GetMailboxIDFromRemoteID(ctx, op.client, mboxID)
	})
}

func (op *EntOpsRead) GetMailboxName(ctx context.Context, mboxID imap.InternalMailboxID) (string, error) {
	return wrapEntErrFnTyped(func() (string, error) {
		return GetMailboxName(ctx, op.client, mboxID)
	})
}

func (op *EntOpsRead) GetMailboxNameWithRemoteID(ctx context.Context, mboxID imap.MailboxID) (string, error) {
	return wrapEntErrFnTyped(func() (string, error) {
		return GetMailboxNameWithRemoteID(ctx, op.client, mboxID)
	})
}

func (op *EntOpsRead) GetMailboxMessageIDPairs(ctx context.Context, mboxID imap.InternalMailboxID) ([]db.MessageIDPair, error) {
	return wrapEntErrFnTyped(func() ([]db.MessageIDPair, error) {
		return GetMailboxMessageIDPairs(ctx, op.client, mboxID)
	})
}

func (op *EntOpsRead) GetAllMailboxesWithAttr(ctx context.Context) ([]*db.Mailbox, error) {
	return wrapEntErrFnTyped(func() ([]*db.Mailbox, error) {
		val, err := GetAllMailboxes(ctx, op.client)

		return xslices.Map(val, entMBoxToDB), err
	})
}

func (op *EntOpsRead) GetAllMailboxesAsRemoteIDs(ctx context.Context) ([]imap.MailboxID, error) {
	return wrapEntErrFnTyped(func() ([]imap.MailboxID, error) {
		val, err := GetAllMailboxes(ctx, op.client)

		return xslices.Map(val, func(t *internal.Mailbox) imap.MailboxID {
			return t.RemoteID
		}), err
	})
}

func (op *EntOpsRead) GetMailboxByName(ctx context.Context, name string) (*db.Mailbox, error) {
	return wrapEntErrFnTyped(func() (*db.Mailbox, error) {
		val, err := GetMailboxByName(ctx, op.client, name)

		return entMBoxToDB(val), err
	})
}

func (op *EntOpsRead) GetMailboxByID(ctx context.Context, mboxID imap.InternalMailboxID) (*db.Mailbox, error) {
	return wrapEntErrFnTyped(func() (*db.Mailbox, error) {
		val, err := GetMailboxByID(ctx, op.client, mboxID)

		return entMBoxToDB(val), err
	})
}

func (op *EntOpsRead) GetMailboxByRemoteID(ctx context.Context, mboxID imap.MailboxID) (*db.Mailbox, error) {
	return wrapEntErrFnTyped(func() (*db.Mailbox, error) {
		val, err := GetMailboxByRemoteID(ctx, op.client, mboxID)

		return entMBoxToDB(val), err
	})
}

func (op *EntOpsRead) GetMailboxRecentCount(ctx context.Context, mboxID imap.InternalMailboxID) (int, error) {
	return wrapEntErrFnTyped(func() (int, error) {
		mbox, err := GetMailboxByID(ctx, op.client, mboxID)
		if err != nil {
			return 0, err
		}

		return GetMailboxRecentCount(ctx, op.client, mbox)
	})
}

func (op *EntOpsRead) GetMailboxMessageCount(ctx context.Context, mboxID imap.InternalMailboxID) (int, error) {
	return wrapEntErrFnTyped(func() (int, error) {
		return GetMailboxMessageCount(ctx, op.client, mboxID)
	})
}

func (op *EntOpsRead) GetMailboxMessageCountWithRemoteID(ctx context.Context, mboxID imap.MailboxID) (int, error) {
	return wrapEntErrFnTyped(func() (int, error) {
		mbox, err := GetMailboxByRemoteID(ctx, op.client, mboxID)
		if err != nil {
			return 0, err
		}

		return GetMailboxMessageCount(ctx, op.client, mbox.ID)
	})
}

func (op *EntOpsRead) GetMailboxFlags(ctx context.Context, mboxID imap.InternalMailboxID) (imap.FlagSet, error) {
	return wrapEntErrFnTyped(func() (imap.FlagSet, error) {
		return GetMailboxFlags(ctx, op.client, mboxID)
	})
}

func (op *EntOpsRead) GetMailboxPermanentFlags(ctx context.Context, mboxID imap.InternalMailboxID) (imap.FlagSet, error) {
	return wrapEntErrFnTyped(func() (imap.FlagSet, error) {
		return GetMailboxPermanentFlags(ctx, op.client, mboxID)
	})
}

func (op *EntOpsRead) GetMailboxAttributes(ctx context.Context, mboxID imap.InternalMailboxID) (imap.FlagSet, error) {
	return wrapEntErrFnTyped(func() (imap.FlagSet, error) {
		return GetMailboxAttributes(ctx, op.client, mboxID)
	})
}

func (op *EntOpsRead) GetMailboxUID(ctx context.Context, mboxID imap.InternalMailboxID) (imap.UID, error) {
	return wrapEntErrFnTyped(func() (imap.UID, error) {
		return GetMailboxUID(ctx, op.client, mboxID)
	})
}

func (op *EntOpsRead) GetMailboxMessageCountAndUID(ctx context.Context, mboxID imap.InternalMailboxID) (int, imap.UID, error) {
	var count int

	var uid imap.UID

	err := wrapEntErrFn(func() error {
		var err error

		count, uid, err = GetMailboxMessageCountAndUID(ctx, op.client, mboxID)

		return err
	})

	return count, uid, err
}

func (op *EntOpsRead) GetMailboxMessageForNewSnapshot(ctx context.Context, mboxID imap.InternalMailboxID) ([]db.SnapshotMessageResult, error) {
	return wrapEntErrFnTyped(func() ([]db.SnapshotMessageResult, error) {
		return GetMailboxMessagesForNewSnapshot(ctx, op.client, mboxID)
	})
}

func (op *EntOpsRead) MailboxTranslateRemoteIDs(ctx context.Context, mboxIDs []imap.MailboxID) ([]imap.InternalMailboxID, error) {
	return wrapEntErrFnTyped(func() ([]imap.InternalMailboxID, error) {
		return TranslateRemoteMailboxIDs(ctx, op.client, mboxIDs)
	})
}

func (op *EntOpsRead) MailboxFilterContains(ctx context.Context, mboxID imap.InternalMailboxID, messageIDs []db.MessageIDPair) ([]imap.InternalMessageID, error) {
	return wrapEntErrFnTyped(func() ([]imap.InternalMessageID, error) {
		return FilterMailboxContains(ctx, op.client, mboxID, messageIDs)
	})
}

func (op *EntOpsRead) MailboxFilterContainsInternalID(ctx context.Context, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) ([]imap.InternalMessageID, error) {
	return wrapEntErrFnTyped(func() ([]imap.InternalMessageID, error) {
		return FilterMailboxContainsInternalID(ctx, op.client, mboxID, messageIDs)
	})
}

func (op *EntOpsRead) GetMailboxCount(ctx context.Context) (int, error) {
	return wrapEntErrFnTyped(func() (int, error) {
		return GetMailboxCount(ctx, op.client)
	})
}

func (op *EntOpsRead) GetMailboxMessageUIDsWithFlagsAfterAddOrUIDBump(ctx context.Context, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) ([]db.UIDWithFlags, error) {
	return wrapEntErrFnTyped(func() ([]db.UIDWithFlags, error) {
		return GetMessageUIDsWithFlagsAfterAddOrUIDBump(ctx, op.client, mboxID, messageIDs)
	})
}

func (op *EntOpsRead) MessageExists(ctx context.Context, id imap.InternalMessageID) (bool, error) {
	return wrapEntErrFnTyped(func() (bool, error) {
		return HasMessageWithID(ctx, op.client, id)
	})
}

func (op *EntOpsRead) MessageExistsWithRemoteID(ctx context.Context, id imap.MessageID) (bool, error) {
	return wrapEntErrFnTyped(func() (bool, error) {
		return HasMessageWithRemoteID(ctx, op.client, id)
	})
}

func (op *EntOpsRead) GetMessageNoEdges(ctx context.Context, id imap.InternalMessageID) (*db.Message, error) {
	return wrapEntErrFnTyped(func() (*db.Message, error) {
		msg, err := GetMessage(ctx, op.client, id)

		return entMessageToDB(msg), err
	})
}

func (op *EntOpsRead) GetTotalMessageCount(ctx context.Context) (int, error) {
	return wrapEntErrFnTyped(func() (int, error) {
		return op.client.Message.Query().Count(ctx)
	})
}

func (op *EntOpsRead) GetMessageRemoteID(ctx context.Context, id imap.InternalMessageID) (imap.MessageID, error) {
	return wrapEntErrFnTyped(func() (imap.MessageID, error) {
		return GetMessageRemoteIDFromID(ctx, op.client, id)
	})
}

func (op *EntOpsRead) GetImportedMessageData(ctx context.Context, id imap.InternalMessageID) (*db.Message, error) {
	return wrapEntErrFnTyped(func() (*db.Message, error) {
		msg, err := GetImportedMessageData(ctx, op.client, id)

		return entMessageToDB(msg), err
	})
}

func (op *EntOpsRead) GetMessageDateAndSize(ctx context.Context, id imap.InternalMessageID) (time.Time, int, error) {
	var date time.Time

	var size int

	err := wrapEntErrFn(func() error {
		msg, err := GetMessageDateAndSize(ctx, op.client, id)
		if err != nil {
			return err
		}

		date = msg.Date
		size = msg.Size

		return err
	})

	return date, size, err
}

func (op *EntOpsRead) GetMessageMailboxIDs(ctx context.Context, id imap.InternalMessageID) ([]imap.InternalMailboxID, error) {
	return wrapEntErrFnTyped(func() ([]imap.InternalMailboxID, error) {
		return GetMessageMailboxIDs(ctx, op.client, id)
	})
}

func (op *EntOpsRead) GetMessagesFlags(ctx context.Context, ids []imap.InternalMessageID) ([]db.MessageFlagSet, error) {
	return wrapEntErrFnTyped(func() ([]db.MessageFlagSet, error) {
		return GetMessageFlags(ctx, op.client, ids)
	})
}

func (op *EntOpsRead) GetMessageIDsMarkedAsDelete(ctx context.Context) ([]imap.InternalMessageID, error) {
	return wrapEntErrFnTyped(func() ([]imap.InternalMessageID, error) {
		return GetMessageIDsMarkedDeleted(ctx, op.client)
	})
}

func (op *EntOpsRead) GetMessageIDFromRemoteID(ctx context.Context, id imap.MessageID) (imap.InternalMessageID, error) {
	return wrapEntErrFnTyped(func() (imap.InternalMessageID, error) {
		return GetMessageIDFromRemoteID(ctx, op.client, id)
	})
}

func (op *EntOpsRead) GetMessageDeletedFlag(ctx context.Context, id imap.InternalMessageID) (bool, error) {
	return wrapEntErrFnTyped(func() (bool, error) {
		msg, err := GetMessageWithIDWithDeletedFlag(ctx, op.client, id)
		if err != nil {
			return false, err
		}

		return msg.Deleted, nil
	})
}

func (op *EntOpsRead) GetAllMessagesIDsAsMap(ctx context.Context) (map[imap.InternalMessageID]struct{}, error) {
	return wrapEntErrFnTyped(func() (map[imap.InternalMessageID]struct{}, error) {
		return GetAllMessagesIDsAsMap(ctx, op.client)
	})
}

func (op *EntOpsRead) GetDeletedSubscriptionSet(ctx context.Context) (map[imap.MailboxID]*db.DeletedSubscription, error) {
	return wrapEntErrFnTyped(func() (map[imap.MailboxID]*db.DeletedSubscription, error) {
		ent, err := GetDeletedSubscriptionSet(ctx, op.client)
		if err != nil {
			return nil, err
		}

		result := make(map[imap.MailboxID]*db.DeletedSubscription, len(ent))

		for k, v := range ent {
			result[k] = entSubscriptionToDB(v)
		}

		return result, nil
	})
}
