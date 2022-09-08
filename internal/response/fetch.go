package response

import (
	"fmt"

	"github.com/ProtonMail/gluon/imap"
)

type fetch struct {
	seq   imap.SeqID
	items []Item
}

func Fetch(seq imap.SeqID) *fetch {
	return &fetch{
		seq: seq,
	}
}

func (r *fetch) WithItems(items ...Item) *fetch {
	r.items = append(r.items, items...)
	return r
}

func (r *fetch) Send(s Session) error {
	return s.WriteResponse(r.String())
}

func (r *fetch) String() string {
	var items []string

	for _, item := range r.items {
		items = append(items, item.String())
	}

	return fmt.Sprintf(`* %v FETCH (%v)`, r.seq, join(items))
}
