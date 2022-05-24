package imap

import (
	"fmt"

	"github.com/ProtonMail/gluon/internal/utils"
)

type MailboxIDChanged struct {
	*updateWaiter

	OldID string
	NewID string
}

func NewMailboxIDChanged(oldID, newID string) MailboxIDChanged {
	return MailboxIDChanged{
		updateWaiter: newUpdateWaiter(),
		OldID:        oldID,
		NewID:        newID,
	}
}

func (u MailboxIDChanged) String() string {
	return fmt.Sprintf("MailboxIDChanged: OldID = %v, NewID = %v", utils.ShortID(u.OldID), utils.ShortID(u.NewID))
}

func (MailboxIDChanged) _isUpdate() {}
