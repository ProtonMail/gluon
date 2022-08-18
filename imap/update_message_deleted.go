package imap

import (
	"fmt"
)

type MessageDeleted struct {
	*updateWaiter

	MessageID MessageID
}

func NewMessagesDeleted(messageID MessageID) *MessageDeleted {
	return &MessageDeleted{
		updateWaiter: newUpdateWaiter(),
		MessageID:    messageID,
	}
}

func (u *MessageDeleted) String() string {
	return fmt.Sprintf("MessageDeleted ID=%v", u.MessageID.ShortID())
}

func (u *MessageDeleted) _isUpdate() {}
