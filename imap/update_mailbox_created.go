package imap

import (
	"fmt"
	"strings"

	"github.com/ProtonMail/gluon/internal/utils"
)

type MailboxCreated struct {
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
		utils.ShortID(u.Mailbox.ID),
		utils.ShortID(strings.Join(u.Mailbox.Name, "/")),
	)
}

func (*MailboxCreated) _isUpdate() {}
