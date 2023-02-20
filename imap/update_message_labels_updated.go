package imap

import (
	"fmt"

	"github.com/bradenaw/juniper/xslices"
)

type MessageMailboxesUpdated struct {
	updateBase

	*updateWaiter

	MessageID            MessageID
	MailboxIDs           []MailboxID
	Seen, Flagged, Draft bool
}

func NewMessageMailboxesUpdated(messageID MessageID, mailboxIDs []MailboxID, seen, flagged, draft bool) *MessageMailboxesUpdated {
	return &MessageMailboxesUpdated{
		updateWaiter: newUpdateWaiter(),
		MessageID:    messageID,
		MailboxIDs:   mailboxIDs,
		Seen:         seen,
		Flagged:      flagged,
		Draft:        draft,
	}
}

func (u *MessageMailboxesUpdated) String() string {
	return fmt.Sprintf(
		"MessageMailboxesUpdated: MessageID = %v, MailboxIDs = %v, seen = %v, flagged = %v, draft = %v",
		u.MessageID.ShortID(),
		xslices.Map(u.MailboxIDs, func(id MailboxID) string { return id.ShortID() }),
		u.Seen,
		u.Flagged,
		u.Draft,
	)
}
