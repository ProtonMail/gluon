package imap

import (
	"fmt"
)

type MailboxIDChanged struct {
	updateBase

	*updateWaiter

	InternalID InternalMailboxID
	RemoteID   MailboxID
}

func NewMailboxIDChanged(internalID InternalMailboxID, remoteID MailboxID) *MailboxIDChanged {
	return &MailboxIDChanged{
		updateWaiter: newUpdateWaiter(),
		InternalID:   internalID,
		RemoteID:     remoteID,
	}
}

func (u *MailboxIDChanged) String() string {
	return fmt.Sprintf("MailboxIDChanged: InternalID = %v, RemoteID = %v", u.InternalID.ShortID(), u.RemoteID.ShortID())
}
