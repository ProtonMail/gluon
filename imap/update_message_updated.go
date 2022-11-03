package imap

import (
	"fmt"
	"github.com/bradenaw/juniper/xslices"
)

// MessageUpdated replaces the previous behavior of MessageDelete followed by MessageCreate. Furthermore, it guarantees
// that the operation is executed atomically.
type MessageUpdated struct {
	updateBase
	*updateWaiter

	Message       Message
	Literal       []byte
	MailboxIDs    []MailboxID
	ParsedMessage *ParsedMessage
}

func NewMessageUpdated(message Message, literal []byte, mailboxIDs []MailboxID, parsedMessage *ParsedMessage) *MessageUpdated {
	return &MessageUpdated{
		updateWaiter:  newUpdateWaiter(),
		Message:       message,
		Literal:       literal,
		MailboxIDs:    mailboxIDs,
		ParsedMessage: parsedMessage,
	}
}

func (u *MessageUpdated) String() string {
	return fmt.Sprintf("MessageUpdate: ID:%v Mailboxes:%v Flags:%s",
		u.Message.ID.ShortID(),
		xslices.Map(u.MailboxIDs, func(mboxID MailboxID) string {
			return mboxID.ShortID()
		}),
		u.Message.Flags.ToSlice(),
	)
}
