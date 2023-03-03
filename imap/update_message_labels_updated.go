package imap

import (
	"fmt"

	"github.com/bradenaw/juniper/xslices"
)

type MessageMailboxesUpdated struct {
	updateBase

	*updateWaiter

	MessageID   MessageID
	MailboxIDs  []MailboxID
	CustomFlags MessageCustomFlags
}

type MessageCustomFlags struct {
	Seen, Flagged, Draft, Answered bool
}

func (f MessageCustomFlags) String() string {
	return fmt.Sprintf("seen=%v, flagged=%v, draft=%v, answered=%v", f.Seen, f.Flagged, f.Draft, f.Answered)
}

func NewMessageMailboxesUpdated(messageID MessageID, mailboxIDs []MailboxID, flags MessageCustomFlags) *MessageMailboxesUpdated {
	return &MessageMailboxesUpdated{
		updateWaiter: newUpdateWaiter(),
		MessageID:    messageID,
		MailboxIDs:   mailboxIDs,
		CustomFlags:  flags,
	}
}

func (u *MessageMailboxesUpdated) String() string {
	return fmt.Sprintf(
		"MessageMailboxesUpdated: MessageID = %v, MailboxIDs = %v, CustomFlags = %v",
		u.MessageID.ShortID(),
		xslices.Map(u.MailboxIDs, func(id MailboxID) string { return id.ShortID() }),
		u.CustomFlags,
	)
}
