package imap

type Noop struct {
	updateBase

	*updateWaiter
}

func NewNoop() *Noop {
	return &Noop{
		updateWaiter: newUpdateWaiter(),
	}
}

func (u *Noop) String() string {
	return "Noop"
}
