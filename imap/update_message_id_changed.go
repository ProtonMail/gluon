package imap

import (
	"fmt"
)

type MessageIDChanged struct {
	*updateWaiter

	InternalID InternalMessageID
	RemoteID   MessageID
}

func NewMessageIDChanged(internalID InternalMessageID, remoteID MessageID) *MessageIDChanged {
	return &MessageIDChanged{
		updateWaiter: newUpdateWaiter(),
		InternalID:   internalID,
		RemoteID:     remoteID,
	}
}

func (u *MessageIDChanged) String() string {
	return fmt.Sprintf("MessageID changed: InternalID = %v, RemoteID = %v", u.InternalID.ShortID(), u.RemoteID.ShortID())
}

func (*MessageIDChanged) _isUpdate() {}
