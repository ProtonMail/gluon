package utils

import (
	"context"
	"time"

	"github.com/ProtonMail/gluon/db"
	"github.com/ProtonMail/gluon/imap"
	"github.com/sirupsen/logrus"
)

// ReadTracer prints all method names to a trace log.
type ReadTracer struct {
	RD    db.ReadOnly
	Entry *logrus.Entry
}

func (r ReadTracer) MailboxExistsWithID(ctx context.Context, mboxID imap.InternalMailboxID) (bool, error) {
	r.Entry.Tracef("MailboxExistsWithID")

	return r.RD.MailboxExistsWithID(ctx, mboxID)
}

func (r ReadTracer) MailboxExistsWithRemoteID(ctx context.Context, mboxID imap.MailboxID) (bool, error) {
	r.Entry.Tracef("MailboxExistsWithRemoteID")

	return r.RD.MailboxExistsWithRemoteID(ctx, mboxID)
}

func (r ReadTracer) MailboxExistsWithName(ctx context.Context, name string) (bool, error) {
	r.Entry.Tracef("MailboxExistsWithName")

	return r.RD.MailboxExistsWithName(ctx, name)
}

func (r ReadTracer) GetMailboxIDFromRemoteID(ctx context.Context, mboxID imap.MailboxID) (imap.InternalMailboxID, error) {
	r.Entry.Tracef("GetMailboxIDFromRemoteID")

	return r.RD.GetMailboxIDFromRemoteID(ctx, mboxID)
}

func (r ReadTracer) GetMailboxName(ctx context.Context, mboxID imap.InternalMailboxID) (string, error) {
	r.Entry.Tracef("GetMailboxName")

	return r.RD.GetMailboxName(ctx, mboxID)
}

func (r ReadTracer) GetMailboxNameWithRemoteID(ctx context.Context, mboxID imap.MailboxID) (string, error) {
	r.Entry.Tracef("GetMailboxNameWithRemoteID")

	return r.RD.GetMailboxNameWithRemoteID(ctx, mboxID)
}

func (r ReadTracer) GetMailboxMessageIDPairs(ctx context.Context, mboxID imap.InternalMailboxID) ([]db.MessageIDPair, error) {
	r.Entry.Tracef("GetMailboxMessageIDPairs")

	return r.RD.GetMailboxMessageIDPairs(ctx, mboxID)
}

func (r ReadTracer) GetAllMailboxesWithAttr(ctx context.Context) ([]*db.MailboxWithAttr, error) {
	r.Entry.Tracef("GetAllMailboxesWithAttr")

	return r.RD.GetAllMailboxesWithAttr(ctx)
}

func (r ReadTracer) GetAllMailboxesAsRemoteIDs(ctx context.Context) ([]imap.MailboxID, error) {
	r.Entry.Tracef("GetAllMailboxesAsRemoteIDs")

	return r.RD.GetAllMailboxesAsRemoteIDs(ctx)
}

func (r ReadTracer) GetMailboxByName(ctx context.Context, name string) (*db.Mailbox, error) {
	r.Entry.Tracef("GetMailboxByName")

	return r.RD.GetMailboxByName(ctx, name)
}

func (r ReadTracer) GetMailboxByID(ctx context.Context, mboxID imap.InternalMailboxID) (*db.Mailbox, error) {
	r.Entry.Tracef("GetMailboxByID")

	return r.RD.GetMailboxByID(ctx, mboxID)
}

func (r ReadTracer) GetMailboxByRemoteID(ctx context.Context, mboxID imap.MailboxID) (*db.Mailbox, error) {
	r.Entry.Tracef("GetMailboxByRemoteID")

	return r.RD.GetMailboxByRemoteID(ctx, mboxID)
}

func (r ReadTracer) GetMailboxRecentCount(ctx context.Context, mboxID imap.InternalMailboxID) (int, error) {
	r.Entry.Tracef("GetMailboxRecentCount")

	return r.RD.GetMailboxRecentCount(ctx, mboxID)
}

func (r ReadTracer) GetMailboxMessageCount(ctx context.Context, mboxID imap.InternalMailboxID) (int, error) {
	r.Entry.Tracef("GetMailboxMessageCount")

	return r.RD.GetMailboxMessageCount(ctx, mboxID)
}

func (r ReadTracer) GetMailboxMessageCountWithRemoteID(ctx context.Context, mboxID imap.MailboxID) (int, error) {
	r.Entry.Tracef("GetMailboxMessageCountWithRemoteID")

	return r.RD.GetMailboxMessageCountWithRemoteID(ctx, mboxID)
}

func (r ReadTracer) GetMailboxFlags(ctx context.Context, mboxID imap.InternalMailboxID) (imap.FlagSet, error) {
	r.Entry.Tracef("GetMailboxFlags")

	return r.RD.GetMailboxFlags(ctx, mboxID)
}

func (r ReadTracer) GetMailboxPermanentFlags(ctx context.Context, mboxID imap.InternalMailboxID) (imap.FlagSet, error) {
	r.Entry.Tracef("GetMailboxPermanentFlags")

	return r.RD.GetMailboxPermanentFlags(ctx, mboxID)
}

func (r ReadTracer) GetMailboxAttributes(ctx context.Context, mboxID imap.InternalMailboxID) (imap.FlagSet, error) {
	r.Entry.Tracef("GetMailboxAttributes")

	return r.RD.GetMailboxAttributes(ctx, mboxID)
}

func (r ReadTracer) GetMailboxUID(ctx context.Context, mboxID imap.InternalMailboxID) (imap.UID, error) {
	r.Entry.Tracef("GetMailboxUID")

	return r.RD.GetMailboxUID(ctx, mboxID)
}

func (r ReadTracer) GetMailboxMessageCountAndUID(ctx context.Context, mboxID imap.InternalMailboxID) (int, imap.UID, error) {
	r.Entry.Tracef("GetMailboxMessageCountAndUID")

	return r.RD.GetMailboxMessageCountAndUID(ctx, mboxID)
}

func (r ReadTracer) GetMailboxMessageForNewSnapshot(ctx context.Context, mboxID imap.InternalMailboxID) ([]db.SnapshotMessageResult, error) {
	r.Entry.Tracef("GetMailboxMessagesForNewSnapshot")

	return r.RD.GetMailboxMessageForNewSnapshot(ctx, mboxID)
}

func (r ReadTracer) MailboxTranslateRemoteIDs(ctx context.Context, mboxIDs []imap.MailboxID) ([]imap.InternalMailboxID, error) {
	r.Entry.Tracef("MailboxTranslateRemoteIDs")

	return r.RD.MailboxTranslateRemoteIDs(ctx, mboxIDs)
}

func (r ReadTracer) MailboxFilterContains(ctx context.Context, mboxID imap.InternalMailboxID, messageIDs []db.MessageIDPair) ([]imap.InternalMessageID, error) {
	r.Entry.Tracef("MailboxFilterContains")

	return r.RD.MailboxFilterContains(ctx, mboxID, messageIDs)
}

func (r ReadTracer) GetMailboxCount(ctx context.Context) (int, error) {
	r.Entry.Tracef("GetMailboxCount")

	return r.RD.GetMailboxCount(ctx)
}

func (r ReadTracer) MessageExists(ctx context.Context, id imap.InternalMessageID) (bool, error) {
	r.Entry.Tracef("MessageExists")

	return r.RD.MessageExists(ctx, id)
}

func (r ReadTracer) MessageExistsWithRemoteID(ctx context.Context, id imap.MessageID) (bool, error) {
	r.Entry.Tracef("MessageExistsWithRemoteID")

	return r.RD.MessageExistsWithRemoteID(ctx, id)
}

func (r ReadTracer) GetMessageNoEdges(ctx context.Context, id imap.InternalMessageID) (*db.Message, error) {
	r.Entry.Tracef("GetMessagesNoEdges")

	return r.RD.GetMessageNoEdges(ctx, id)
}

func (r ReadTracer) GetTotalMessageCount(ctx context.Context) (int, error) {
	r.Entry.Tracef("GetTotalMessagecount")

	return r.RD.GetTotalMessageCount(ctx)
}

func (r ReadTracer) GetMessageRemoteID(ctx context.Context, id imap.InternalMessageID) (imap.MessageID, error) {
	r.Entry.Tracef("GetMessageRemoteID")

	return r.RD.GetMessageRemoteID(ctx, id)
}

func (r ReadTracer) GetImportedMessageData(ctx context.Context, id imap.InternalMessageID) (*db.MessageWithFlags, error) {
	r.Entry.Tracef("GetImportedMessageData")

	return r.RD.GetImportedMessageData(ctx, id)
}

func (r ReadTracer) GetMessageDateAndSize(ctx context.Context, id imap.InternalMessageID) (time.Time, int, error) {
	r.Entry.Tracef("GetMessageDateAndSize")

	return r.RD.GetMessageDateAndSize(ctx, id)
}

func (r ReadTracer) GetMessageMailboxIDs(ctx context.Context, id imap.InternalMessageID) ([]imap.InternalMailboxID, error) {
	r.Entry.Tracef("GetMessageMailboxIDs")

	return r.RD.GetMessageMailboxIDs(ctx, id)
}

func (r ReadTracer) GetMessagesFlags(ctx context.Context, ids []imap.InternalMessageID) ([]db.MessageFlagSet, error) {
	r.Entry.Tracef("GetMessageFlags")

	return r.RD.GetMessagesFlags(ctx, ids)
}

func (r ReadTracer) GetMessageIDsMarkedAsDelete(ctx context.Context) ([]imap.InternalMessageID, error) {
	r.Entry.Tracef("GetMessageIDsMarkedAsDelete")

	return r.RD.GetMessageIDsMarkedAsDelete(ctx)
}

func (r ReadTracer) GetMessageIDFromRemoteID(ctx context.Context, id imap.MessageID) (imap.InternalMessageID, error) {
	r.Entry.Tracef("GetMessageIDFromRemoteID")

	return r.RD.GetMessageIDFromRemoteID(ctx, id)
}

func (r ReadTracer) GetMessageDeletedFlag(ctx context.Context, id imap.InternalMessageID) (bool, error) {
	r.Entry.Tracef("GetMessageDeletedFlag")

	return r.RD.GetMessageDeletedFlag(ctx, id)
}

func (r ReadTracer) GetAllMessagesIDsAsMap(ctx context.Context) (map[imap.InternalMessageID]struct{}, error) {
	r.Entry.Tracef("GetAllMessagesIDsAsMap")

	return r.RD.GetAllMessagesIDsAsMap(ctx)
}

func (r ReadTracer) GetDeletedSubscriptionSet(ctx context.Context) (map[imap.MailboxID]*db.DeletedSubscription, error) {
	r.Entry.Tracef("GetDeletedSubscriptionSet")

	return r.RD.GetDeletedSubscriptionSet(ctx)
}

func (r ReadTracer) GetConnectorSettings(ctx context.Context) (string, bool, error) {
	r.Entry.Tracef("GetConnectorSettings")

	return r.RD.GetConnectorSettings(ctx)
}

// WriteTracer prints all method names to a trace log.
type WriteTracer struct {
	ReadTracer
	TX db.Transaction
}

func (w WriteTracer) CreateMailbox(
	ctx context.Context,
	mboxID imap.MailboxID,
	name string,
	flags, permFlags, attrs imap.FlagSet,
	uidValidity imap.UID,
) (*db.Mailbox, error) {
	w.Entry.Tracef("CreateMailbox")

	return w.TX.CreateMailbox(ctx, mboxID, name, flags, permFlags, attrs, uidValidity)
}

func (w WriteTracer) GetOrCreateMailbox(
	ctx context.Context,
	mboxID imap.MailboxID,
	name string,
	flags, permFlags, attrs imap.FlagSet,
	uidValidity imap.UID,
) (*db.Mailbox, error) {
	w.Entry.Tracef("GetOrCreateMailbox")

	return w.TX.GetOrCreateMailbox(ctx, mboxID, name, flags, permFlags, attrs, uidValidity)
}

func (w WriteTracer) GetOrCreateMailboxAlt(
	ctx context.Context,
	mbox imap.Mailbox,
	delimiter string,
	uidValidity imap.UID,
) (*db.Mailbox, error) {
	w.Entry.Tracef("GetOrCreateMailboxAlt")

	return w.TX.GetOrCreateMailboxAlt(ctx, mbox, delimiter, uidValidity)
}

func (w WriteTracer) RenameMailboxWithRemoteID(ctx context.Context, mboxID imap.MailboxID, name string) error {
	w.Entry.Tracef("RenameMailboxWithRemoteID")

	return w.TX.RenameMailboxWithRemoteID(ctx, mboxID, name)
}

func (w WriteTracer) DeleteMailboxWithRemoteID(ctx context.Context, mboxID imap.MailboxID) error {
	w.Entry.Tracef("DeleteMailboxWithRemoteID")

	return w.TX.DeleteMailboxWithRemoteID(ctx, mboxID)
}

func (w WriteTracer) AddMessagesToMailbox(
	ctx context.Context,
	mboxID imap.InternalMailboxID,
	messageIDs []db.MessageIDPair,
) ([]db.UIDWithFlags, error) {
	w.Entry.Tracef("AddMessagesToMailbox")

	return w.TX.AddMessagesToMailbox(ctx, mboxID, messageIDs)
}

func (w WriteTracer) RemoveMessagesFromMailbox(
	ctx context.Context,
	mboxID imap.InternalMailboxID,
	messageIDs []imap.InternalMessageID,
) error {
	w.Entry.Tracef("RemoveMessagesFromMailbox")

	return w.TX.RemoveMessagesFromMailbox(ctx, mboxID, messageIDs)
}

func (w WriteTracer) ClearRecentFlagInMailboxOnMessage(
	ctx context.Context,
	mboxID imap.InternalMailboxID,
	messageID imap.InternalMessageID,
) error {
	w.Entry.Tracef("ClearRecentFlagInMailboxOnMessage")

	return w.TX.ClearRecentFlagInMailboxOnMessage(ctx, mboxID, messageID)
}

func (w WriteTracer) ClearRecentFlagsInMailbox(ctx context.Context, mboxID imap.InternalMailboxID) error {
	w.Entry.Tracef("ClearRecentFlagsInMailbox")

	return w.TX.ClearRecentFlagsInMailbox(ctx, mboxID)
}

func (w WriteTracer) CreateMailboxIfNotExists(ctx context.Context, mbox imap.Mailbox, delimiter string, uidValidity imap.UID) error {
	w.Entry.Tracef("ClearMailboxIfNotExists")

	return w.TX.CreateMailboxIfNotExists(ctx, mbox, delimiter, uidValidity)
}

func (w WriteTracer) SetMailboxMessagesDeletedFlag(
	ctx context.Context,
	mboxID imap.InternalMailboxID,
	messageIDs []imap.InternalMessageID,
	deleted bool,
) error {
	w.Entry.Tracef("SetMailboxMessagesDeleteFlag")

	return w.TX.SetMailboxMessagesDeletedFlag(ctx, mboxID, messageIDs, deleted)
}

func (w WriteTracer) SetMailboxSubscribed(ctx context.Context, mboxID imap.InternalMailboxID, subscribed bool) error {
	w.Entry.Tracef("SetMailboxSubscribed")

	return w.TX.SetMailboxSubscribed(ctx, mboxID, subscribed)
}

func (w WriteTracer) UpdateRemoteMailboxID(ctx context.Context, mobxID imap.InternalMailboxID, remoteID imap.MailboxID) error {
	w.Entry.Tracef("UpdateRemoteMailboxID")

	return w.TX.UpdateRemoteMailboxID(ctx, mobxID, remoteID)
}

func (w WriteTracer) SetMailboxUIDValidity(ctx context.Context, mboxID imap.InternalMailboxID, uidValidity imap.UID) error {
	w.Entry.Tracef("SetMailboxUIDValidity")

	return w.TX.SetMailboxUIDValidity(ctx, mboxID, uidValidity)
}

func (w WriteTracer) CreateMessages(ctx context.Context, reqs ...*db.CreateMessageReq) error {
	w.Entry.Tracef("CreateMessages")

	return w.TX.CreateMessages(ctx, reqs...)
}

func (w WriteTracer) CreateMessageAndAddToMailbox(
	ctx context.Context,
	mbox imap.InternalMailboxID,
	req *db.CreateMessageReq,
) (imap.UID, imap.FlagSet, error) {
	w.Entry.Tracef("CreateMessageAndAddToMailbox")

	return w.TX.CreateMessageAndAddToMailbox(ctx, mbox, req)
}

func (w WriteTracer) MarkMessageAsDeleted(ctx context.Context, id imap.InternalMessageID) error {
	w.Entry.Tracef("MarkMessageAsDeleted")

	return w.TX.MarkMessageAsDeleted(ctx, id)
}

func (w WriteTracer) MarkMessageAsDeletedAndAssignRandomRemoteID(ctx context.Context, id imap.InternalMessageID) error {
	w.Entry.Tracef("MarkMessageAsDeletedAndAssignRandomRemoteID")

	return w.TX.MarkMessageAsDeletedAndAssignRandomRemoteID(ctx, id)
}

func (w WriteTracer) MarkMessageAsDeletedWithRemoteID(ctx context.Context, id imap.MessageID) error {
	w.Entry.Tracef("MarkMessageAsDeletedWithRemoteID")

	return w.TX.MarkMessageAsDeletedWithRemoteID(ctx, id)
}

func (w WriteTracer) DeleteMessages(ctx context.Context, ids []imap.InternalMessageID) error {
	w.Entry.Tracef("DeleteMessages")

	return w.TX.DeleteMessages(ctx, ids)
}

func (w WriteTracer) UpdateRemoteMessageID(ctx context.Context, internalID imap.InternalMessageID, remoteID imap.MessageID) error {
	w.Entry.Tracef("UpdateRemoteMessageID")

	return w.TX.UpdateRemoteMessageID(ctx, internalID, remoteID)
}

func (w WriteTracer) AddFlagToMessages(ctx context.Context, ids []imap.InternalMessageID, flag string) error {
	w.Entry.Tracef("AddFlagsToMessage")

	return w.TX.AddFlagToMessages(ctx, ids, flag)
}

func (w WriteTracer) RemoveFlagFromMessages(ctx context.Context, ids []imap.InternalMessageID, flag string) error {
	w.Entry.Tracef("RemoveFlagsFromMessages")

	return w.TX.RemoveFlagFromMessages(ctx, ids, flag)
}

func (w WriteTracer) SetFlagsOnMessages(ctx context.Context, ids []imap.InternalMessageID, flags imap.FlagSet) error {
	w.Entry.Tracef("SetFlagsOnMessages")

	return w.TX.SetFlagsOnMessages(ctx, ids, flags)
}

func (w WriteTracer) AddDeletedSubscription(ctx context.Context, mboxName string, mboxID imap.MailboxID) error {
	w.Entry.Tracef("AddDeletedSubscription")

	return w.TX.AddDeletedSubscription(ctx, mboxName, mboxID)
}

func (w WriteTracer) RemoveDeletedSubscriptionWithName(ctx context.Context, mboxName string) (int, error) {
	w.Entry.Tracef("RemoveDeletedSubscriptionWithName")

	return w.TX.RemoveDeletedSubscriptionWithName(ctx, mboxName)
}

func (w WriteTracer) StoreConnectorSettings(ctx context.Context, settings string) error {
	w.Entry.Tracef("StoreConnectorSettings")

	return w.TX.StoreConnectorSettings(ctx, settings)
}
