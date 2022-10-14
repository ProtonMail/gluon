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
	return s.WriteResponse(r)
}

func (r *fetch) Strings() (raw string, filtered string) {
	var items, filteredItems []string

	for _, item := range r.items {
		rawStr, filteredStr := item.Strings()
		items = append(items, rawStr)
		filteredItems = append(filteredItems, filteredStr)
	}

	raw = fmt.Sprintf(`* %v FETCH (%v)`, r.seq, join(items))
	filtered = fmt.Sprintf(`* %v FETCH (%v)`, r.seq, join(filteredItems))

	return raw, filtered
}

func (r *fetch) canSkip(other Response) bool {
	otherExists, isExists := other.(*exists)
	if isExists && r.seq < otherExists.count {
		return true
	}

	_, isRecent := other.(*recent)
	if isRecent {
		return true
	}

	otherFetch, isFetch := other.(*fetch)
	if isFetch && otherFetch.seq != r.seq {
		return true
	}

	return false
}

func (r *fetch) mergeWith(other Response) Response {
	otherFetch, ok := other.(*fetch)
	if !ok || otherFetch.seq != r.seq {
		return nil
	}

	for _, item := range r.items {
		otherFetch.items = appendOrMergeItem(otherFetch.items, item)
	}

	return otherFetch
}
