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

func (r *status) String(isPrivateByDefault bool) string {
	var items []string

	for _, item := range r.items {
		items = append(items, item.String(isPrivateByDefault))
	}

	return fmt.Sprintf(`* STATUS %v (%v)`, strconv.Quote(r.name), join(items))
}
