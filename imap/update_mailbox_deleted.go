package imap

import (
	"fmt"
)

type MailboxDeleted struct {
	updateBase

	*updateWaiter

	MailboxID MailboxID
}

func NewMailboxDeleted(mailboxID MailboxID) *MailboxDeleted {
	return &MailboxDeleted{
		updateWaiter: newUpdateWaiter(),
		MailboxID:    mailboxID,
	}
}

func (u *MailboxDeleted) String() string {
	return fmt.Sprintf("MailboxDeleted: MailboxID = %v", u.MailboxID.ShortID())
}
