package imap

import (
	"fmt"

	"github.com/bradenaw/juniper/xslices"
)

type MessageMailboxesUpdated struct {
	updateBase

	*updateWaiter

	MessageID  MessageID
	MailboxIDs []MailboxID
	Flags      FlagSet
}

func NewMessageMailboxesUpdated(messageID MessageID, mailboxIDs []MailboxID, flags FlagSet) *MessageMailboxesUpdated {
	return &MessageMailboxesUpdated{
		updateWaiter: newUpdateWaiter(),
		MessageID:    messageID,
		MailboxIDs:   mailboxIDs,
		Flags:        flags,
	}
}

func (u *MessageMailboxesUpdated) String() string {
	return fmt.Sprintf(
		"MessageMailboxesUpdated: MessageID = %v, MailboxIDs = %v, Flags = %v",
		u.MessageID.ShortID(),
		xslices.Map(u.MailboxIDs, func(id MailboxID) string { return id.ShortID() }),
		u.Flags.ToSlice(),
	)
}
