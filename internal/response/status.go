package response

import (
	"fmt"
	"strconv"
)

type status struct {
	name  string
	items []Item
}

func Status() *status {
	return &status{}
}

func (r *status) WithMailbox(name string) *status {
	r.name = name
	return r
}

func (r *status) WithItems(item ...Item) *status {
	r.items = append(r.items, item...)
	return r
}

func (r *status) Send(s Session) error {
	return s.WriteResponse(r)
}

func (r *status) Strings() (raw string, filtered string) {
	var items, itemsFiltered []string

	for _, item := range r.items {
		rawPart, filteredPart := item.Strings()
		items = append(items, rawPart)
		itemsFiltered = append(itemsFiltered, filteredPart)
	}

	raw = fmt.Sprintf(`* STATUS %v (%v)`, strconv.Quote(r.name), join(items))
	filtered = fmt.Sprintf(`* STATUS %v (%v)`, strconv.Quote(r.name), join(itemsFiltered))

	return raw, filtered
}
