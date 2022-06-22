package imap

import (
	"fmt"

	"github.com/ProtonMail/gluon/internal/utils"
)

type MessageDeleted struct {
	*updateWaiter

	MessageID string
}

func NewMessagesDeleted(messageID string) *MessageDeleted {
	return &MessageDeleted{
		updateWaiter: newUpdateWaiter(),
		MessageID:    messageID,
	}
}

func (u *MessageDeleted) String() string {
	return fmt.Sprintf("MessageDeleted ID=%v", utils.ShortID(u.MessageID))
}

func (u *MessageDeleted) _isUpdate() {}
