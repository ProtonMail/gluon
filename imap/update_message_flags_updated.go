package imap

import (
	"fmt"
)

type MessageFlagsUpdated struct {
	updateBase

	*updateWaiter

	MessageID     MessageID
	Seen, Flagged bool
}

func NewMessageFlagsUpdated(messageID MessageID, seen, flagged bool) *MessageFlagsUpdated {
	return &MessageFlagsUpdated{
		updateWaiter: newUpdateWaiter(),
		MessageID:    messageID,
		Seen:         seen,
		Flagged:      flagged,
	}
}

func (u *MessageFlagsUpdated) String() string {
	return fmt.Sprintf(
		"MessageFlagsUpdated: MessageID = %v, seen = %v, flagged = %v",
		u.MessageID.ShortID(),
		u.Seen,
		u.Flagged,
	)
}
