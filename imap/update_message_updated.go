package imap

import (
	"fmt"

	"github.com/bradenaw/juniper/xslices"
)

// MessageUpdated replaces the previous behavior of MessageDelete followed by MessageCreate.
// If the message does exist, it is updated.
// If the message does not exist, it can optionally be created.
// Furthermore, it guarantees that the operation is executed atomically.
type MessageUpdated struct {
	updateBase
	*updateWaiter

	Message       Message
	Literal       []byte
	MailboxIDs    []MailboxID
	ParsedMessage *ParsedMessage
	AllowCreate   bool
}

func NewMessageUpdated(
	message Message,
	literal []byte,
	mailboxIDs []MailboxID,
	parsedMessage *ParsedMessage,
	allowCreate bool,
) *MessageUpdated {
	return &MessageUpdated{
		updateWaiter:  newUpdateWaiter(),
		Message:       message,
		Literal:       literal,
		MailboxIDs:    mailboxIDs,
		ParsedMessage: parsedMessage,
		AllowCreate:   allowCreate,
	}
}

func (u *MessageUpdated) String() string {
	return fmt.Sprintf("MessageUpdated: ID:%v Mailboxes:%v Flags:%s AllowCreate:%v",
		u.Message.ID.ShortID(),
		xslices.Map(u.MailboxIDs, func(mboxID MailboxID) string {
			return mboxID.ShortID()
		}),
		u.Message.Flags.ToSlice(),
		u.AllowCreate,
	)
}
