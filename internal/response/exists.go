package response

import (
	"fmt"

	"github.com/ProtonMail/gluon/imap"
)

type exists struct {
	count imap.SeqID
}

func Exists() *exists {
	return &exists{}
}

func (r *exists) WithCount(n imap.SeqID) *exists {
	r.count = n
	return r
}

func (r *exists) Send(s Session) error {
	return s.WriteResponse(r)
}

func (r *exists) String(_ bool) string {
	return fmt.Sprintf("* %v EXISTS", r.count)
}

func (r *exists) canSkip(other Response) bool {
	if _, isRecent := other.(*recent); isRecent {
		return true
	}

	if _, isFetch := other.(*fetch); isFetch {
		return true
	}

	return false
}

func (r *exists) mergeWith(other Response) Response {
	otherExist, ok := other.(*exists)
	if !ok {
		return nil
	}

	if otherExist.count > r.count {
		panic(fmt.Sprintf(
			"consecutive exists must be non-decreasing, but had %d and new %d",
			otherExist.count, r.count,
		))
	}

	return r
}
