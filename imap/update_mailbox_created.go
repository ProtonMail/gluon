package imap

import (
	"fmt"
	"strings"

	"github.com/ProtonMail/gluon/internal/utils"
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
		utils.ShortID(strings.Join(u.Mailbox.Name, "/")),
	)
}
