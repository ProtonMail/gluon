package response

import (
	"fmt"
	"strconv"

	"github.com/ProtonMail/gluon/imap"
	"github.com/emersion/go-imap/utf7"
)

type list struct {
	name, del string
	att       imap.FlagSet
}

func List() *list {
	return &list{att: imap.NewFlagSet()}
}

func (r *list) WithName(name string) *list {
	r.name = name
	return r
}

func (r *list) WithDelimiter(del string) *list {
	r.del = del
	return r
}

func (r *list) WithAttributes(att imap.FlagSet) *list {
	r.att = r.att.AddFlagSet(att)
	return r
}

func (r *list) Send(s Session) error {
	return s.WriteResponse(r.String())
}

func (r *list) String() string {
	del := "NIL"

	if r.del != "" {
		del = strconv.Quote(r.del)
	}

	enc, err := utf7.Encoding.NewEncoder().String(r.name)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf(`* LIST (%v) %v %v`, join(r.att.ToSlice()), del, strconv.Quote(enc))
}
