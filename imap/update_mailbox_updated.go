package imap

import (
	"fmt"
	"strings"

	"github.com/ProtonMail/gluon/internal/utils"
)

type MailboxUpdated struct {
	*updateWaiter

	MailboxID   string
	MailboxName []string
}

func NewMailboxUpdated(mailboxID string, mailboxName []string) MailboxUpdated {
	return MailboxUpdated{
		updateWaiter: newUpdateWaiter(),
		MailboxID:    mailboxID,
		MailboxName:  mailboxName,
	}
}

func (u MailboxUpdated) String() string {
	return fmt.Sprintf(
		"MailboxUpdated: MailboxID = %v, MailboxName = %v",
		utils.ShortID(u.MailboxID),
		utils.ShortID(strings.Join(u.MailboxName, "/")),
	)
}

func (MailboxUpdated) _isUpdate() {}
