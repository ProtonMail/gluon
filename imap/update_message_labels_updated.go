package imap

import (
	"fmt"

	"github.com/bradenaw/juniper/xslices"
)

type MessageLabelsUpdated struct {
	*updateWaiter

	MessageID     MessageID
	MailboxIDs    []LabelID
	Seen, Flagged bool
}

func NewMessageLabelsUpdated(messageID MessageID, mailboxIDs []LabelID, seen, flagged bool) *MessageLabelsUpdated {
	return &MessageLabelsUpdated{
		updateWaiter: newUpdateWaiter(),
		MessageID:    messageID,
		MailboxIDs:   mailboxIDs,
		Seen:         seen,
		Flagged:      flagged,
	}
}

func (u *MessageLabelsUpdated) String() string {
	return fmt.Sprintf(
		"MessageLabelsUpdated: MessageID = %v, MailboxIDs = %v, seen = %v, flagged = %v",
		u.MessageID.ShortID(),
		xslices.Map(u.MailboxIDs, func(id LabelID) string { return id.ShortID() }),
		u.Seen,
		u.Flagged,
	)
}

func (*MessageLabelsUpdated) _isUpdate() {}
