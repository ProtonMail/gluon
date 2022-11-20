package imap

type UIDValidityBumped struct {
	updateBase

	*updateWaiter
}

func NewUIDValidityBumped() *UIDValidityBumped {
	return &UIDValidityBumped{
		updateWaiter: newUpdateWaiter(),
	}
}

func (u *UIDValidityBumped) String() string {
	return "UIDValidityBumped"
}
