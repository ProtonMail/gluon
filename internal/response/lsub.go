package response

import (
	"fmt"
	"strconv"

	"github.com/ProtonMail/gluon/imap"
)

type lsub struct {
	name, del string
	att       imap.FlagSet
}

func Lsub() *lsub {
	return &lsub{att: imap.NewFlagSet()}
}

func (r *lsub) WithName(name string) *lsub {
	r.name = name
	return r
}

func (r *lsub) WithDelimiter(del string) *lsub {
	r.del = del
	return r
}

func (r *lsub) WithAttributes(att imap.FlagSet) *lsub {
	r.att.AddFlagSetToSelf(att)
	return r
}

func (r *lsub) Send(s Session) error {
	return s.WriteResponse(r)
}

func (r *lsub) Strings() (raw string, _ string) {
	del := "NIL"

	if r.del != "" {
		del = strconv.Quote(r.del)
	}

	raw = fmt.Sprintf(`* LSUB (%v) %v %v`, join(r.att.ToSlice()), del, strconv.Quote(r.name))

	return raw, raw
}
