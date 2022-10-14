package response

import (
	"fmt"

	"github.com/ProtonMail/gluon/imap"
)

type expunge struct {
	seq imap.SeqID
}

func Expunge(seq imap.SeqID) *expunge {
	return &expunge{
		seq: seq,
	}
}

func (r *expunge) Send(s Session) error {
	return s.WriteResponse(r)
}

func (r *expunge) String(_ bool) string {
	return fmt.Sprintf("* %v EXPUNGE", r.seq)
}
