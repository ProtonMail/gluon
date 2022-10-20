package imap

import (
	"fmt"

	"github.com/bradenaw/juniper/xslices"
)

type MessageMailboxesUpdated struct {
	updateBase

	*updateWaiter

	MessageID     MessageID
	MailboxIDs    []MailboxID
	Seen, Flagged bool
}

func NewMessageMailboxesUpdated(messageID MessageID, mailboxIDs []MailboxID, seen, flagged bool) *MessageMailboxesUpdated {
	return &MessageMailboxesUpdated{
		updateWaiter: newUpdateWaiter(),
		MessageID:    messageID,
		MailboxIDs:   mailboxIDs,
		Seen:         seen,
		Flagged:      flagged,
	}
}

func (u *MessageMailboxesUpdated) String() string {
	return fmt.Sprintf(
		"MessageMailboxesUpdated: MessageID = %v, MailboxIDs = %v, seen = %v, flagged = %v",
		u.MessageID.ShortID(),
		xslices.Map(u.MailboxIDs, func(id MailboxID) string { return id.ShortID() }),
		u.Seen,
		u.Flagged,
	)
}
