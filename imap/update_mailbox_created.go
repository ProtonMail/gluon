package imap

import (
	"fmt"
	"strings"
)

type MailboxCreated struct {
	updateBase

	*updateWaiter

	Mailbox Mailbox
}

func NewMailboxCreated(mailbox Mailbox) *MailboxCreated {
	return &MailboxCreated{
		updateWaiter: newUpdateWaiter(),
		Mailbox:      mailbox,
	}
}

func (u *MailboxCreated) String() string {
	return fmt.Sprintf(
		"MailboxCreated: Mailbox.ID = %v, Mailbox.Name = %v",
		u.Mailbox.ID.ShortID(),
		ShortID(strings.Join(u.Mailbox.Name, "/")),
	)
}
