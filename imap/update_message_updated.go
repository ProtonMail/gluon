package imap

import (
	"fmt"

	"github.com/bradenaw/juniper/xslices"
)

type MessageUpdated struct {
	*updateWaiter

	MessageID     MessageID
	MailboxIDs    []LabelID
	Seen, Flagged bool
}

func NewMessageUpdated(messageID MessageID, mailboxIDs []LabelID, seen, flagged bool) *MessageUpdated {
	return &MessageUpdated{
		updateWaiter: newUpdateWaiter(),
		MessageID:    messageID,
		MailboxIDs:   mailboxIDs,
		Seen:         seen,
		Flagged:      flagged,
	}
}

func (u *MessageUpdated) String() string {
	return fmt.Sprintf(
		"MessageUpdated: MessageID = %v, MailboxIDs = %v, seen = %v, flagged = %v",
		u.MessageID.ShortID(),
		xslices.Map(u.MailboxIDs, func(id LabelID) string { return id.ShortID() }),
		u.Seen,
		u.Flagged,
	)
}

func (*MessageUpdated) _isUpdate() {}
