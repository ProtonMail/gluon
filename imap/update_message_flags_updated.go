package imap

import (
	"fmt"
)

type MessageFlagsUpdated struct {
	updateBase

	*updateWaiter

	MessageID   MessageID
	CustomFlags MessageCustomFlags
}

func NewMessageFlagsUpdated(messageID MessageID, flags MessageCustomFlags) *MessageFlagsUpdated {
	return &MessageFlagsUpdated{
		updateWaiter: newUpdateWaiter(),
		MessageID:    messageID,
		CustomFlags:  flags,
	}
}

func (u *MessageFlagsUpdated) String() string {
	return fmt.Sprintf(
		"MessageFlagsUpdated: MessageID = %v, CustomFlags = %v",
		u.MessageID.ShortID(),
		u.CustomFlags,
	)
}
