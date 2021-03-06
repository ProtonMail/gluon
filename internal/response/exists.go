package response

import (
	"fmt"
)

type exists struct {
	count int
}

func Exists() *exists {
	return &exists{}
}

func (r *exists) WithCount(n int) *exists {
	r.count = n
	return r
}

func (r *exists) Send(s Session) error {
	return s.WriteResponse(r.String())
}

func (r *exists) String() string {
	return fmt.Sprintf("* %v EXISTS", r.count)
}
