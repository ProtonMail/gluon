package imap

import "fmt"

type UIDValidityBumped struct {
	updateBase

	*updateWaiter

	UIDValidity UID
}

func NewUIDValidityBumped(validity UID) *UIDValidityBumped {
	return &UIDValidityBumped{
		updateWaiter: newUpdateWaiter(),

		UIDValidity: validity,
	}
}

func (u *UIDValidityBumped) String() string {
	return fmt.Sprintf("UIDValidityBumped: %v", u.UIDValidity)
}
