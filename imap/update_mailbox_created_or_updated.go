package imap

import (
	"fmt"
	"strings"
)

type MailboxUpdatedOrCreated struct {
	updateBase

	*updateWaiter

	Mailbox Mailbox
}

func NewMailboxUpdatedOrCreated(mailbox Mailbox) *MailboxUpdatedOrCreated {
	return &MailboxUpdatedOrCreated{
		updateWaiter: newUpdateWaiter(),
		Mailbox:      mailbox,
	}
}

func (u *MailboxUpdatedOrCreated) String() string {
	return fmt.Sprintf("MailboxUpdatedOrCreated: Mailbox.ID = %v, Mailbox.Name = %v",
		u.Mailbox.ID.ShortID(),
		ShortID(strings.Join(u.Mailbox.Name, "/")),
	)
}
