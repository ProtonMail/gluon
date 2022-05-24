package response

import (
	"fmt"
)

type expunge struct {
	seq int
}

func Expunge(seq int) *expunge {
	return &expunge{
		seq: seq,
	}
}

func (r *expunge) Send(s Session) error {
	return s.WriteResponse(r.String())
}

func (r *expunge) String() string {
	return fmt.Sprintf("* %v EXPUNGE", r.seq)
}
