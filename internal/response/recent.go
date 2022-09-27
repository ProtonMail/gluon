package response

import (
	"fmt"
)

type recent struct {
	count uint32
}

func isRecent(r Response) bool {
	_, ok := r.(*recent)
	return ok
}

func recentHasHigherID(a, b Response) bool {
	recentA, ok := a.(*recent)
	if !ok {
		return false
	}

	recentB, ok := b.(*recent)
	if !ok {
		return false
	}

	return recentA.count > recentB.count
}

func Recent() *recent {
	return &recent{}
}

func (r *recent) WithCount(n uint32) *recent {
	r.count = n
	return r
}

func (r *recent) Send(s Session) error {
	return s.WriteResponse(r.String())
}

func (r *recent) String() string {
	return fmt.Sprintf("* %v RECENT", r.count)
}
