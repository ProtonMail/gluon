package imap

import (
	"fmt"

	"github.com/ProtonMail/gluon/internal/utils"
)

type MailboxDeleted struct {
	*updateWaiter

	MailboxID string
}

func NewMailboxDeleted(mailboxID string) *MailboxDeleted {
	return &MailboxDeleted{
		updateWaiter: newUpdateWaiter(),
		MailboxID:    mailboxID,
	}
}

func (u *MailboxDeleted) String() string {
	return fmt.Sprintf("MailboxDeleted: MailboxID = %v", utils.ShortID(u.MailboxID))
}

func (*MailboxDeleted) _isUpdate() {}
