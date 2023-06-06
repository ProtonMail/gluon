package sqlite3

import (
	"context"
	"time"

	"github.com/ProtonMail/gluon/db"
	"github.com/ProtonMail/gluon/imap"
	"github.com/sirupsen/logrus"
)

// ReadTracer prints all method names to a trace log.
type ReadTracer struct {
	rd    db.ReadOnly
	entry *logrus.Entry
}

func (r ReadTracer) MailboxExistsWithID(ctx context.Context, mboxID imap.InternalMailboxID) (bool, error) {
	r.entry.Tracef("MailboxExistsWithID")

	return r.rd.MailboxExistsWithID(ctx, mboxID)
}

func (r ReadTracer) MailboxExistsWithRemoteID(ctx context.Context, mboxID imap.MailboxID) (bool, error) {
	r.entry.Tracef("MailboxExistsWithRemoteID")

	return r.rd.MailboxExistsWithRemoteID(ctx, mboxID)
}

func (r ReadTracer) MailboxExistsWithName(ctx context.Context, name string) (bool, error) {
	r.entry.Tracef("MailboxExistsWithName")

	return r.rd.MailboxExistsWithName(ctx, name)
}

func (r ReadTracer) GetMailboxIDFromRemoteID(ctx context.Context, mboxID imap.MailboxID) (imap.InternalMailboxID, error) {
	r.entry.Tracef("GetMailboxIDFromRemoteID")

	return r.rd.GetMailboxIDFromRemoteID(ctx, mboxID)
}

func (r ReadTracer) GetMailboxName(ctx context.Context, mboxID imap.InternalMailboxID) (string, error) {
	r.entry.Tracef("GetMailboxName")

	return r.rd.GetMailboxName(ctx, mboxID)
}

func (r ReadTracer) GetMailboxNameWithRemoteID(ctx context.Context, mboxID imap.MailboxID) (string, error) {
	r.entry.Tracef("GetMailboxNameWithRemoteID")

	return r.rd.GetMailboxNameWithRemoteID(ctx, mboxID)
}

func (r ReadTracer) GetMailboxMessageIDPairs(ctx context.Context, mboxID imap.InternalMailboxID) ([]db.MessageIDPair, error) {
	r.entry.Tracef("GetMailboxMessageIDPairs")

	return r.rd.GetMailboxMessageIDPairs(ctx, mboxID)
}

func (r ReadTracer) GetAllMailboxesWithAttr(ctx context.Context) ([]*db.Mailbox, error) {
	r.entry.Tracef("GetAllMailboxesWithAttr")

	return r.rd.GetAllMailboxesWithAttr(ctx)
}

func (r ReadTracer) GetAllMailboxesAsRemoteIDs(ctx context.Context) ([]imap.MailboxID, error) {
	r.entry.Tracef("GetAllMailboxesAsRemoteIDs")

	return r.rd.GetAllMailboxesAsRemoteIDs(ctx)
}

func (r ReadTracer) GetMailboxByName(ctx context.Context, name string) (*db.Mailbox, error) {
	r.entry.Tracef("GetMailboxByName")

	return r.rd.GetMailboxByName(ctx, name)
}

func (r ReadTracer) GetMailboxByID(ctx context.Context, mboxID imap.InternalMailboxID) (*db.Mailbox, error) {
	r.entry.Tracef("GetMailboxByID")

	return r.rd.GetMailboxByID(ctx, mboxID)
}

func (r ReadTracer) GetMailboxByRemoteID(ctx context.Context, mboxID imap.MailboxID) (*db.Mailbox, error) {
	r.entry.Tracef("GetMailboxByRemoteID")

	return r.rd.GetMailboxByRemoteID(ctx, mboxID)
}

func (r ReadTracer) GetMailboxRecentCount(ctx context.Context, mboxID imap.InternalMailboxID) (int, error) {
	r.entry.Tracef("GetMailboxRecentCount")

	return r.rd.GetMailboxRecentCount(ctx, mboxID)
}

func (r ReadTracer) GetMailboxMessageCount(ctx context.Context, mboxID imap.InternalMailboxID) (int, error) {
	r.entry.Tracef("GetMailboxMessageCount")

	return r.rd.GetMailboxMessageCount(ctx, mboxID)
}

func (r ReadTracer) GetMailboxMessageCountWithRemoteID(ctx context.Context, mboxID imap.MailboxID) (int, error) {
	r.entry.Tracef("GetMailboxMessageCountWithRemoteID")

	return r.rd.GetMailboxMessageCountWithRemoteID(ctx, mboxID)
}

func (r ReadTracer) GetMailboxFlags(ctx context.Context, mboxID imap.InternalMailboxID) (imap.FlagSet, error) {
	r.entry.Tracef("GetMailboxFlags")

	return r.rd.GetMailboxFlags(ctx, mboxID)
}

func (r ReadTracer) GetMailboxPermanentFlags(ctx context.Context, mboxID imap.InternalMailboxID) (imap.FlagSet, error) {
	r.entry.Tracef("GetMailboxPermanentFlags")

	return r.rd.GetMailboxPermanentFlags(ctx, mboxID)
}

func (r ReadTracer) GetMailboxAttributes(ctx context.Context, mboxID imap.InternalMailboxID) (imap.FlagSet, error) {
	r.entry.Tracef("GetMailboxAttributes")

	return r.rd.GetMailboxAttributes(ctx, mboxID)
}

func (r ReadTracer) GetMailboxUID(ctx context.Context, mboxID imap.InternalMailboxID) (imap.UID, error) {
	r.entry.Tracef("GetMailboxUID")

	return r.rd.GetMailboxUID(ctx, mboxID)
}

func (r ReadTracer) GetMailboxMessageCountAndUID(ctx context.Context, mboxID imap.InternalMailboxID) (int, imap.UID, error) {
	r.entry.Tracef("GetMailboxMessageCountAndUID")

	return r.rd.GetMailboxMessageCountAndUID(ctx, mboxID)
}

func (r ReadTracer) GetMailboxMessageForNewSnapshot(ctx context.Context, mboxID imap.InternalMailboxID) ([]db.SnapshotMessageResult, error) {
	r.entry.Tracef("GetMailboxMessagesForNewSnapshot")

	return r.rd.GetMailboxMessageForNewSnapshot(ctx, mboxID)
}

func (r ReadTracer) MailboxTranslateRemoteIDs(ctx context.Context, mboxIDs []imap.MailboxID) ([]imap.InternalMailboxID, error) {
	r.entry.Tracef("MailboxTranslateRemoteIDs")

	return r.rd.MailboxTranslateRemoteIDs(ctx, mboxIDs)
}

func (r ReadTracer) MailboxFilterContains(ctx context.Context, mboxID imap.InternalMailboxID, messageIDs []db.MessageIDPair) ([]imap.InternalMessageID, error) {
	r.entry.Tracef("MailboxFilterContains")

	return r.rd.MailboxFilterContains(ctx, mboxID, messageIDs)
}

func (r ReadTracer) MailboxFilterContainsInternalID(ctx context.Context, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) ([]imap.InternalMessageID, error) {
	r.entry.Tracef("MailboxFilterContainsInternalID")

	return r.rd.MailboxFilterContainsInternalID(ctx, mboxID, messageIDs)
}

func (r ReadTracer) GetMailboxCount(ctx context.Context) (int, error) {
	r.entry.Tracef("GetMailboxCount")

	return r.rd.GetMailboxCount(ctx)
}

func (r ReadTracer) GetMailboxMessageUIDsWithFlagsAfterAddOrUIDBump(ctx context.Context, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) ([]db.UIDWithFlags, error) {
	r.entry.Tracef("GetMailboxMessageUIDsWithFlagsAfterAddOrUIDBump")

	return r.rd.GetMailboxMessageUIDsWithFlagsAfterAddOrUIDBump(ctx, mboxID, messageIDs)
}

func (r ReadTracer) MessageExists(ctx context.Context, id imap.InternalMessageID) (bool, error) {
	r.entry.Tracef("MessageExists")

	return r.rd.MessageExists(ctx, id)
}

func (r ReadTracer) MessageExistsWithRemoteID(ctx context.Context, id imap.MessageID) (bool, error) {
	r.entry.Tracef("MessageExistsWithRemoteID")

	return r.rd.MessageExistsWithRemoteID(ctx, id)
}

func (r ReadTracer) GetMessageNoEdges(ctx context.Context, id imap.InternalMessageID) (*db.Message, error) {
	r.entry.Tracef("GetMessagesNoEdges")

	return r.rd.GetMessageNoEdges(ctx, id)
}

func (r ReadTracer) GetTotalMessageCount(ctx context.Context) (int, error) {
	r.entry.Tracef("GetTotalMessagecount")

	return r.rd.GetTotalMessageCount(ctx)
}

func (r ReadTracer) GetMessageRemoteID(ctx context.Context, id imap.InternalMessageID) (imap.MessageID, error) {
	r.entry.Tracef("GetMessageRemoteID")

	return r.rd.GetMessageRemoteID(ctx, id)
}

func (r ReadTracer) GetImportedMessageData(ctx context.Context, id imap.InternalMessageID) (*db.Message, error) {
	r.entry.Tracef("GetImportedMessageData")

	return r.rd.GetImportedMessageData(ctx, id)
}

func (r ReadTracer) GetMessageDateAndSize(ctx context.Context, id imap.InternalMessageID) (time.Time, int, error) {
	r.entry.Tracef("GetMessageDateAndSize")

	return r.rd.GetMessageDateAndSize(ctx, id)
}

func (r ReadTracer) GetMessageMailboxIDs(ctx context.Context, id imap.InternalMessageID) ([]imap.InternalMailboxID, error) {
	r.entry.Tracef("GetMailboxIDs")

	return r.rd.GetMessageMailboxIDs(ctx, id)
}

func (r ReadTracer) GetMessagesFlags(ctx context.Context, ids []imap.InternalMessageID) ([]db.MessageFlagSet, error) {
	r.entry.Tracef("GetMessageFlags")

	return r.rd.GetMessagesFlags(ctx, ids)
}

func (r ReadTracer) GetMessageIDsMarkedAsDelete(ctx context.Context) ([]imap.InternalMessageID, error) {
	r.entry.Tracef("GetMessageIDsMarkedAsDelete")

	return r.rd.GetMessageIDsMarkedAsDelete(ctx)
}

func (r ReadTracer) GetMessageIDFromRemoteID(ctx context.Context, id imap.MessageID) (imap.InternalMessageID, error) {
	r.entry.Tracef("GetMessageIDFromRemoteID")

	return r.rd.GetMessageIDFromRemoteID(ctx, id)
}

func (r ReadTracer) GetMessageDeletedFlag(ctx context.Context, id imap.InternalMessageID) (bool, error) {
	r.entry.Tracef("GetMessageDeletedFlag")

	return r.rd.GetMessageDeletedFlag(ctx, id)
}

func (r ReadTracer) GetAllMessagesIDsAsMap(ctx context.Context) (map[imap.InternalMessageID]struct{}, error) {
	r.entry.Tracef("GetAllMessagesIDsAsMap")

	return r.rd.GetAllMessagesIDsAsMap(ctx)
}

func (r ReadTracer) GetDeletedSubscriptionSet(ctx context.Context) (map[imap.MailboxID]*db.DeletedSubscription, error) {
	r.entry.Tracef("GetDeletedSubscriptionSet")

	return r.rd.GetDeletedSubscriptionSet(ctx)
}

// WriteTracer prints all method names to a trace log.
type WriteTracer struct {
	ReadTracer
	tx db.Transaction
}

func (w WriteTracer) CreateMailbox(
	ctx context.Context,
	mboxID imap.MailboxID,
	name string,
	flags, permFlags, attrs imap.FlagSet,
	uidValidity imap.UID,
) (*db.Mailbox, error) {
	w.entry.Tracef("CreateMailbox")

	return w.tx.CreateMailbox(ctx, mboxID, name, flags, permFlags, attrs, uidValidity)
}

func (w WriteTracer) GetOrCreateMailbox(
	ctx context.Context,
	mboxID imap.MailboxID,
	name string,
	flags, permFlags, attrs imap.FlagSet,
	uidValidity imap.UID,
) (*db.Mailbox, error) {
	w.entry.Tracef("GetOrCreateMailbox")

	return w.tx.GetOrCreateMailbox(ctx, mboxID, name, flags, permFlags, attrs, uidValidity)
}

func (w WriteTracer) GetOrCreateMailboxAlt(
	ctx context.Context,
	mbox imap.Mailbox,
	delimiter string,
	uidValidity imap.UID,
) (*db.Mailbox, error) {
	w.entry.Tracef("GetOrCreateMailboxAlt")

	return w.tx.GetOrCreateMailboxAlt(ctx, mbox, delimiter, uidValidity)
}

func (w WriteTracer) RenameMailboxWithRemoteID(ctx context.Context, mboxID imap.MailboxID, name string) error {
	w.entry.Tracef("RenameMailboxWithRemoteID")

	return w.tx.RenameMailboxWithRemoteID(ctx, mboxID, name)
}

func (w WriteTracer) DeleteMailboxWithRemoteID(ctx context.Context, mboxID imap.MailboxID) error {
	w.entry.Tracef("DeleteMailboxWithRemoteID")

	return w.tx.DeleteMailboxWithRemoteID(ctx, mboxID)
}

func (w WriteTracer) BumpMailboxUIDNext(ctx context.Context, mboxID imap.InternalMailboxID, count int) error {
	w.entry.Tracef("BumpMailboxUIDNext")

	return w.tx.BumpMailboxUIDNext(ctx, mboxID, count)
}

func (w WriteTracer) AddMessagesToMailbox(
	ctx context.Context,
	mboxID imap.InternalMailboxID,
	messageIDs []imap.InternalMessageID,
) ([]db.UIDWithFlags, error) {
	w.entry.Tracef("AddMessagesToMailbox")

	return w.tx.AddMessagesToMailbox(ctx, mboxID, messageIDs)
}

func (w WriteTracer) BumpMailboxUIDsForMessage(
	ctx context.Context,
	mboxID imap.InternalMailboxID,
	messageIDs []imap.InternalMessageID,
) ([]db.UIDWithFlags, error) {
	w.entry.Tracef("BumpMailboxUIDsForMessage")

	return w.tx.BumpMailboxUIDsForMessage(ctx, mboxID, messageIDs)
}

func (w WriteTracer) RemoveMessagesFromMailbox(
	ctx context.Context,
	mboxID imap.InternalMailboxID,
	messageIDs []imap.InternalMessageID,
) error {
	w.entry.Tracef("RemoveMessagesFromMailbox")

	return w.tx.RemoveMessagesFromMailbox(ctx, mboxID, messageIDs)
}

func (w WriteTracer) ClearRecentFlagInMailboxOnMessage(
	ctx context.Context,
	mboxID imap.InternalMailboxID,
	messageID imap.InternalMessageID,
) error {
	w.entry.Tracef("ClearRecentFlagInMailboxOnMessage")

	return w.tx.ClearRecentFlagInMailboxOnMessage(ctx, mboxID, messageID)
}

func (w WriteTracer) ClearRecentFlagsInMailbox(ctx context.Context, mboxID imap.InternalMailboxID) error {
	w.entry.Tracef("ClearRecentFlagsInMailbox")

	return w.tx.ClearRecentFlagsInMailbox(ctx, mboxID)
}

func (w WriteTracer) CreateMailboxIfNotExists(ctx context.Context, mbox imap.Mailbox, delimiter string, uidValidity imap.UID) error {
	w.entry.Tracef("ClearMailboxIfNotExists")

	return w.tx.CreateMailboxIfNotExists(ctx, mbox, delimiter, uidValidity)
}

func (w WriteTracer) SetMailboxMessagesDeletedFlag(
	ctx context.Context,
	mboxID imap.InternalMailboxID,
	messageIDs []imap.InternalMessageID,
	deleted bool,
) error {
	w.entry.Tracef("SetMailboxMessagesDeleteFlag")

	return w.tx.SetMailboxMessagesDeletedFlag(ctx, mboxID, messageIDs, deleted)
}

func (w WriteTracer) SetMailboxSubscribed(ctx context.Context, mboxID imap.InternalMailboxID, subscribed bool) error {
	w.entry.Tracef("SetMailboxSubscribed")

	return w.tx.SetMailboxSubscribed(ctx, mboxID, subscribed)
}

func (w WriteTracer) UpdateRemoteMailboxID(ctx context.Context, mobxID imap.InternalMailboxID, remoteID imap.MailboxID) error {
	w.entry.Tracef("UpdateRemoteMailboxID")

	return w.tx.UpdateRemoteMailboxID(ctx, mobxID, remoteID)
}

func (w WriteTracer) SetMailboxUIDValidity(ctx context.Context, mboxID imap.InternalMailboxID, uidValidity imap.UID) error {
	w.entry.Tracef("SetMailboxUIDValidity")

	return w.tx.SetMailboxUIDValidity(ctx, mboxID, uidValidity)
}

func (w WriteTracer) CreateMessages(ctx context.Context, reqs ...*db.CreateMessageReq) ([]*db.Message, error) {
	w.entry.Tracef("CreateMessages")

	return w.tx.CreateMessages(ctx, reqs...)
}

func (w WriteTracer) CreateMessageAndAddToMailbox(
	ctx context.Context,
	mbox imap.InternalMailboxID,
	req *db.CreateMessageReq,
) (imap.UID, imap.FlagSet, error) {
	w.entry.Tracef("CreateMessageAndAddToMailbox")

	return w.tx.CreateMessageAndAddToMailbox(ctx, mbox, req)
}

func (w WriteTracer) MarkMessageAsDeleted(ctx context.Context, id imap.InternalMessageID) error {
	w.entry.Tracef("MarkMessageAsDeleted")

	return w.tx.MarkMessageAsDeleted(ctx, id)
}

func (w WriteTracer) MarkMessageAsDeletedAndAssignRandomRemoteID(ctx context.Context, id imap.InternalMessageID) error {
	w.entry.Tracef("MarkMessageAsDeletedAndAssignRandomRemoteID")

	return w.tx.MarkMessageAsDeletedAndAssignRandomRemoteID(ctx, id)
}

func (w WriteTracer) MarkMessageAsDeletedWithRemoteID(ctx context.Context, id imap.MessageID) error {
	w.entry.Tracef("MarkMessageAsDeletedWithRemoteID")

	return w.tx.MarkMessageAsDeletedWithRemoteID(ctx, id)
}

func (w WriteTracer) DeleteMessages(ctx context.Context, ids []imap.InternalMessageID) error {
	w.entry.Tracef("DeleteMessages")

	return w.tx.DeleteMessages(ctx, ids)
}

func (w WriteTracer) UpdateRemoteMessageID(ctx context.Context, internalID imap.InternalMessageID, remoteID imap.MessageID) error {
	w.entry.Tracef("UpdateRemoteMessageID")

	return w.tx.UpdateRemoteMessageID(ctx, internalID, remoteID)
}

func (w WriteTracer) AddFlagToMessages(ctx context.Context, ids []imap.InternalMessageID, flag string) error {
	w.entry.Tracef("AddFlagsToMessage")

	return w.tx.AddFlagToMessages(ctx, ids, flag)
}

func (w WriteTracer) RemoveFlagFromMessages(ctx context.Context, ids []imap.InternalMessageID, flag string) error {
	w.entry.Tracef("RemoveFlagsFromMessages")

	return w.tx.RemoveFlagFromMessages(ctx, ids, flag)
}

func (w WriteTracer) SetFlagsOnMessages(ctx context.Context, ids []imap.InternalMessageID, flags imap.FlagSet) error {
	w.entry.Tracef("SetFlagsOnMessages")

	return w.tx.SetFlagsOnMessages(ctx, ids, flags)
}

func (w WriteTracer) AddDeletedSubscription(ctx context.Context, mboxName string, mboxID imap.MailboxID) error {
	w.entry.Tracef("AddDeletedSubscription")

	return w.tx.AddDeletedSubscription(ctx, mboxName, mboxID)
}

func (w WriteTracer) RemoveDeletedSubscriptionWithName(ctx context.Context, mboxName string) (int, error) {
	w.entry.Tracef("RemoveDeletedSubscriptionWithName")

	return w.tx.RemoveDeletedSubscriptionWithName(ctx, mboxName)
}
