package imap

import (
	"fmt"
	"strings"

	"github.com/ProtonMail/gluon/internal/utils"
)

type MailboxUpdated struct {
	updateBase

	*updateWaiter

	MailboxID   LabelID
	MailboxName []string
}

func NewMailboxUpdated(mailboxID LabelID, mailboxName []string) *MailboxUpdated {
	return &MailboxUpdated{
		updateWaiter: newUpdateWaiter(),
		MailboxID:    mailboxID,
		MailboxName:  mailboxName,
	}
}

func (u *MailboxUpdated) String() string {
	return fmt.Sprintf(
		"MailboxUpdated: MailboxID = %v, MailboxName = %v",
		u.MailboxID.ShortID(),
		utils.ShortID(strings.Join(u.MailboxName, "/")),
	)
}
