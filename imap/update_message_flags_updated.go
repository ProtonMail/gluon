package imap

import (
	"fmt"
)

type MessageFlagsUpdated struct {
	updateBase

	*updateWaiter

	MessageID            MessageID
	Seen, Flagged, Draft bool
}

func NewMessageFlagsUpdated(messageID MessageID, seen, flagged, draft bool) *MessageFlagsUpdated {
	return &MessageFlagsUpdated{
		updateWaiter: newUpdateWaiter(),
		MessageID:    messageID,
		Seen:         seen,
		Flagged:      flagged,
		Draft:        draft,
	}
}

func (u *MessageFlagsUpdated) String() string {
	return fmt.Sprintf(
		"MessageFlagsUpdated: MessageID = %v, seen = %v, flagged = %v, draft = %v",
		u.MessageID.ShortID(),
		u.Seen,
		u.Flagged,
		u.Draft,
	)
}
