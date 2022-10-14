package response

import (
	"fmt"
)

type ok struct {
	tag   string
	msg   string
	items []Item
}

func Ok(withTag ...string) *ok {
	var tag string

	if len(withTag) > 0 {
		tag = withTag[0]
	} else {
		tag = "*"
	}

	return &ok{
		tag: tag,
	}
}

func (r *ok) WithMessage(msg string) *ok {
	r.msg = msg
	return r
}

func (r *ok) WithItems(items ...Item) *ok {
	r.items = append(r.items, items...)
	return r
}

func (r *ok) Send(s Session) error {
	return s.WriteResponse(r)
}

func (r *ok) Strings() (raw string, filtered string) {
	parts := []string{r.tag, "OK"}
	partsFiltered := []string{r.tag, "OK"}

	if len(r.items) > 0 {
		var items, itemsFiltered []string

		for _, item := range r.items {
			rawPart, filteredPart := item.Strings()
			items = append(items, rawPart)
			itemsFiltered = append(itemsFiltered, filteredPart)
		}

		parts = append(parts, fmt.Sprintf("[%v]", join(items)))
		partsFiltered = append(partsFiltered, fmt.Sprintf("[%v]", join(itemsFiltered)))
	}

	if r.msg != "" {
		parts = append(parts, r.msg)
		partsFiltered = append(partsFiltered, r.msg)
	}

	return join(parts), join(partsFiltered)
}
