package imap

import (
	"fmt"
	"github.com/bradenaw/juniper/xslices"
)

type Update interface {
	Waiter

	String() string

	_isUpdate()
}

type updateBase struct{}

func (updateBase) _isUpdate() {}

// BatchUpdate allows the submission of multiple individual updates as a single update.
type BatchUpdate struct {
	updateBase
	*updateWaiter

	Updates []Update
}

func NewBatchUpdate(updates ...Update) *BatchUpdate {
	return &BatchUpdate{
		updateWaiter: newUpdateWaiter(),
		Updates:      updates,
	}
}

func (u *BatchUpdate) String() string {
	return fmt.Sprintf("BatchUpdate: Updates = %v",
		xslices.Map(u.Updates, func(u Update) string {
			return u.String()
		}),
	)
}
