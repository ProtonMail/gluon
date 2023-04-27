package imap

import (
	"fmt"
)

type MessageFlagsUpdated struct {
	updateBase

	*updateWaiter

	MessageID MessageID
	Flags     FlagSet
}

func NewMessageFlagsUpdated(messageID MessageID, flags FlagSet) *MessageFlagsUpdated {
	return &MessageFlagsUpdated{
		updateWaiter: newUpdateWaiter(),
		MessageID:    messageID,
		Flags:        flags,
	}
}

func (u *MessageFlagsUpdated) String() string {
	return fmt.Sprintf(
		"MessageFlagsUpdated: MessageID = %v, Flags = %v",
		u.MessageID.ShortID(),
		u.Flags.ToSlice(),
	)
}
