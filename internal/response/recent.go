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
	return s.WriteResponse(r.String())
}

func (r *recent) String() string {
	return fmt.Sprintf("* %v RECENT", r.count)
}
