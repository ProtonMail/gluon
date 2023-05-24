package db

import (
	"context"
	"github.com/ProtonMail/gluon/imap"
	"strings"
)

type MailboxReadOps interface {
	MailboxExistsWithID(ctx context.Context, mboxID imap.InternalMailboxID) (bool, error)

	MailboxExistsWithRemoteID(ctx context.Context, mboxID imap.MailboxID) (bool, error)

	MailboxExistsWithName(ctx context.Context, name string) (bool, error)

	GetMailboxIDFromRemoteID(ctx context.Context, mboxID imap.MailboxID) (imap.InternalMailboxID, error)

	GetMailboxName(ctx context.Context, mboxID imap.InternalMailboxID) (string, error)

	GetMailboxNameWithRemoteID(ctx context.Context, mboxID imap.MailboxID) (string, error)

	GetMailboxMessageIDPairs(ctx context.Context, mboxID imap.InternalMailboxID) ([]MessageIDPair, error)

	GetAllMailboxes(ctx context.Context) ([]*Mailbox, error)

	GetAllMailboxesAsRemoteIDs(ctx context.Context) ([]imap.MailboxID, error)

	GetMailboxByName(ctx context.Context, name string) (*Mailbox, error)

	GetMailboxByID(ctx context.Context, mboxID imap.InternalMailboxID) (*Mailbox, error)

	GetMailboxByRemoteID(ctx context.Context, mboxID imap.MailboxID) (*Mailbox, error)

	GetMailboxRecentCount(ctx context.Context, mboxID imap.InternalMailboxID) (int, error)

	GetMailboxMessageCount(ctx context.Context, mboxID imap.InternalMailboxID) (int, error)

	GetMailboxMessageCountWithRemoteID(ctx context.Context, mboxID imap.MailboxID) (int, error)

	GetMailboxFlags(ctx context.Context, mboxID imap.InternalMailboxID) (imap.FlagSet, error)

	GetMailboxPermanentFlags(ctx context.Context, mboxID imap.InternalMailboxID) (imap.FlagSet, error)

	GetMailboxAttributes(ctx context.Context, mboxID imap.InternalMailboxID) (imap.FlagSet, error)

	GetMailboxUID(ctx context.Context, mboxID imap.InternalMailboxID) (imap.UID, error)

	GetMailboxMessageCountAndUID(ctx context.Context, mboxID imap.InternalMailboxID) (int, imap.UID, error)

	GetMailboxMessageForNewSnapshot(ctx context.Context, mboxID imap.InternalMailboxID) ([]SnapshotMessageResult, error)

	MailboxTranslateRemoteIDs(ctx context.Context, mboxIDs []imap.MailboxID) ([]imap.InternalMailboxID, error)

	MailboxFilterContains(ctx context.Context, mboxID imap.InternalMailboxID, messageIDs []MessageIDPair) ([]imap.InternalMessageID, error)

	MailboxFilterContainsInternalID(ctx context.Context, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) ([]imap.InternalMessageID, error)

	GetMailboxCount(ctx context.Context) (int, error)

	// GetMessageUIDsWithFlagsAfterAddOrUIDBump exploits a property of adding a message to or bumping the UIDs of existing message in mailbox. It can only be
	// used if you can guarantee that the messageID list contains only IDs that have recently added or bumped in the mailbox.
	GetMailboxMessageUIDsWithFlagsAfterAddOrUIDBump(ctx context.Context, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) ([]UIDWithFlags, error)
}

type MailboxWriteOps interface {
	MailboxReadOps

	CreateMailbox(ctx context.Context,
		mboxID imap.MailboxID,
		name string,
		flags, permFlags, attrs imap.FlagSet,
		uidValidity imap.UID) (*Mailbox, error)

	GetOrCreateMailbox(ctx context.Context,
		mboxID imap.MailboxID,
		name string,
		flags, permFlags, attrs imap.FlagSet,
		uidValidity imap.UID) (*Mailbox, error)

	GetOrCreateMailboxAlt(ctx context.Context,
		mbox imap.Mailbox,
		delimiter string,
		uidValidity imap.UID) (*Mailbox, error)

	RenameMailboxWithRemoteID(ctx context.Context, mboxID imap.MailboxID, name string) error

	DeleteMailboxWithRemoteID(ctx context.Context, mboxID imap.MailboxID) error

	BumpMailboxUIDNext(ctx context.Context, mboxID imap.InternalMailboxID, count int) error

	AddMessagesToMailbox(ctx context.Context, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) ([]UIDWithFlags, error)

	BumpMailboxUIDsForMessage(ctx context.Context, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) ([]UIDWithFlags, error)

	RemoveMessagesFromMailbox(ctx context.Context, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) error

	ClearRecentFlagInMailboxOnMessage(ctx context.Context, mboxID imap.InternalMailboxID, messageID imap.InternalMessageID) error

	ClearRecentFlagsInMailbox(ctx context.Context, mboxID imap.InternalMailboxID) error

	CreateMailboxIfNotExists(ctx context.Context, mbox imap.Mailbox, delimiter string, uidValidity imap.UID) error

	SetMailboxMessagesDeletedFlag(ctx context.Context, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID, deleted bool) error

	SetMailboxSubscribed(ctx context.Context, mboxID imap.InternalMailboxID, subscribed bool) error

	UpdateRemoteMailboxID(ctx context.Context, mobxID imap.InternalMailboxID, remoteID imap.MailboxID) error

	SetMailboxUIDValidity(ctx context.Context, mboxID imap.InternalMailboxID, uidValidity imap.UID) error
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
		flagSet.AddToSelf(imap.FlagDeleted)
	}

	if msg.Recent {
		flagSet.AddToSelf(imap.FlagRecent)
	}

	return flagSet
}

type UIDWithFlags struct {
	InternalID imap.InternalMessageID `json:"uid_message"`
	RemoteID   imap.MessageID         `json:"remote_id"`
	UID        imap.UID               `json:"uid"`
	Recent     bool                   `json:"recent"`
	Deleted    bool                   `json:"deleted"`
	Flags      string                 `json:"flags"`
}

func (u *UIDWithFlags) GetFlagSet() imap.FlagSet {
	var flagSet imap.FlagSet

	if len(u.Flags) > 0 {
		flags := strings.Split(u.Flags, ",")
		flagSet = imap.NewFlagSetFromSlice(flags)
	} else {
		flagSet = imap.NewFlagSet()
	}

	if u.Deleted {
		flagSet.AddToSelf(imap.FlagDeleted)
	}

	if u.Recent {
		flagSet.AddToSelf(imap.FlagRecent)
	}

	return flagSet
}
