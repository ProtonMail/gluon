package db

import (
	"context"
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/bradenaw/juniper/xslices"
)

type MessageReadOps interface {
	MessageExists(ctx context.Context, id imap.InternalMessageID) (bool, error)

	MessageExistsWithRemoteID(ctx context.Context, id imap.MessageID) (bool, error)

	GetMessageNoEdges(ctx context.Context, id imap.InternalMessageID) (*Message, error)

	GetTotalMessageCount(ctx context.Context) (int, error)

	GetMessageRemoteID(ctx context.Context, id imap.InternalMessageID) (imap.MessageID, error)

	GetImportedMessageData(ctx context.Context, id imap.InternalMessageID) (*Message, error)

	GetMessageDateAndSize(ctx context.Context, id imap.InternalMessageID) (time.Time, int, error)

	GetMessageMailboxIDs(ctx context.Context, id imap.InternalMessageID) ([]imap.InternalMailboxID, error)

	GetMessagesFlags(ctx context.Context, ids []imap.InternalMessageID) ([]MessageFlagSet, error)

	GetMessageIDsMarkedAsDelete(ctx context.Context) ([]imap.InternalMessageID, error)

	GetMessageIDFromRemoteID(ctx context.Context, id imap.MessageID) (imap.InternalMessageID, error)

	GetMessageDeletedFlag(ctx context.Context, id imap.InternalMessageID) (bool, error)

	GetAllMessagesIDsAsMap(ctx context.Context) (map[imap.InternalMessageID]struct{}, error)
}

type MessageWriteOps interface {
	MessageReadOps

	CreateMessages(ctx context.Context, reqs ...*CreateMessageReq) ([]*Message, error)

	CreateMessageAndAddToMailbox(ctx context.Context, mbox imap.InternalMailboxID, req *CreateMessageReq) (imap.UID, imap.FlagSet, error)

	MarkMessageAsDeleted(ctx context.Context, id imap.InternalMessageID) error

	MarkMessageAsDeletedAndAssignRandomRemoteID(ctx context.Context, id imap.InternalMessageID) error

	MarkMessageAsDeletedWithRemoteID(ctx context.Context, id imap.MessageID) error

	DeleteMessages(ctx context.Context, ids []imap.InternalMessageID) error

	UpdateRemoteMessageID(ctx context.Context, internalID imap.InternalMessageID, remoteID imap.MessageID) error

	AddFlagToMessages(ctx context.Context, ids []imap.InternalMessageID, flag string) error

	RemoveFlagFromMessages(ctx context.Context, ids []imap.InternalMessageID, flag string) error

	SetFlagsOnMessages(ctx context.Context, ids []imap.InternalMessageID, flags imap.FlagSet) error
}

type CreateMessageReq struct {
	Message     imap.Message
	InternalID  imap.InternalMessageID
	LiteralSize int
	Body        string
	Structure   string
	Envelope    string
}

type MessageFlagSet struct {
	ID       imap.InternalMessageID
	RemoteID imap.MessageID
	FlagSet  imap.FlagSet
}

func NewFlagSet(msgUID *UID, flags []*MessageFlag) imap.FlagSet {
	flagSet := imap.NewFlagSetFromSlice(xslices.Map(flags, func(flag *MessageFlag) string {
		return flag.Value
	}))

	if msgUID.Deleted {
		flagSet.AddToSelf(imap.FlagDeleted)
	}

	if msgUID.Recent {
		flagSet.AddToSelf(imap.FlagRecent)
	}

	return flagSet
}
