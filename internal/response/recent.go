package response

import (
	"fmt"
)

type recent struct {
	count uint32
}

func Recent() *recent {
	return &recent{}
}

func (r *recent) WithCount(n uint32) *recent {
	r.count = n
	return r
}

func (r *recent) Send(s Session) error {
	return s.WriteResponse(r)
}

func (r *recent) Strings() (raw string, _ string) {
	raw = fmt.Sprintf("* %v RECENT", r.count)
	return raw, raw
}

func (r *recent) canSkip(other Response) bool {
	if _, isExists := other.(*exists); isExists {
		return true
	}

	if _, isFetch := other.(*fetch); isFetch {
		return true
	}

	return false
}

func (r *recent) mergeWith(other Response) Response {
	otherRecent, ok := other.(*recent)
	if !ok {
		return nil
	}

	if otherRecent.count > r.count {
		panic(fmt.Sprintf(
			"consecutive recents must be non-decreasing, but had %d and new %d",
			otherRecent.count, r.count,
		))
	}

	return r
}
