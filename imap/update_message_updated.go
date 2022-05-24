package imap

import (
	"fmt"

	"github.com/ProtonMail/gluon/internal/utils"

	"github.com/bradenaw/juniper/xslices"
)

type MessageUpdated struct {
	*updateWaiter

	MessageID     string
	MailboxIDs    []string
	Seen, Flagged bool
}

func NewMessageUpdated(messageID string, mailboxIDs []string, seen, flagged bool) MessageUpdated {
	return MessageUpdated{
		updateWaiter: newUpdateWaiter(),
		MessageID:    messageID,
		MailboxIDs:   mailboxIDs,
		Seen:         seen,
		Flagged:      flagged,
	}
}

func (u MessageUpdated) String() string {
	return fmt.Sprintf(
		"MessageUpdated: MessageID = %v, MailboxIDs = %v, seen = %v, flagged = %v",
		utils.ShortID(u.MessageID),
		xslices.Map(u.MailboxIDs, func(id string) string { return utils.ShortID(id) }),
		u.Seen,
		u.Flagged,
	)
}

func (MessageUpdated) _isUpdate() {}
