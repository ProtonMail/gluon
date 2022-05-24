package imap

import (
	"fmt"

	"github.com/ProtonMail/gluon/internal/utils"
)

type MessageIDChanged struct {
	*updateWaiter

	OldID string
	NewID string
}

func NewMessageIDChanged(oldID, newID string) MessageIDChanged {
	return MessageIDChanged{
		updateWaiter: newUpdateWaiter(),
		OldID:        oldID,
		NewID:        newID,
	}
}

func (u MessageIDChanged) String() string {
	return fmt.Sprintf("MessageID changed: OldID = %v, NewID = %v", utils.ShortID(u.OldID), utils.ShortID(u.NewID))
}

func (MessageIDChanged) _isUpdate() {}
