package response

import "fmt"

type no struct {
	tag   string
	err   error
	items []Item
}

func No(withTag ...string) *no {
	var tag string

	if len(withTag) > 0 {
		tag = withTag[0]
	} else {
		tag = "*"
	}

	return &no{
		tag: tag,
	}
}

func (r *no) WithItems(items ...Item) *no {
	r.items = append(r.items, items...)
	return r
}

func (r *no) WithError(err error) *no {
	r.err = err
	return r
}

func (r *no) Send(s Session) error {
	return s.WriteResponse(r)
}

func (r *no) Strings() (raw string, filtered string) {
	parts := []string{r.tag, "NO"}
	partsFiltered := []string{r.tag, "NO"}

	if len(r.items) > 0 {
		var items, itemsFiltered []string

		for _, item := range r.items {
			itemRaw, itemFiltered := item.Strings()
			items = append(items, itemRaw)
			itemsFiltered = append(itemsFiltered, itemFiltered)
		}

		parts = append(parts, fmt.Sprintf("[%v]", join(items)))
		partsFiltered = append(partsFiltered, fmt.Sprintf("[%v]", join(itemsFiltered)))
	}

	if r.err != nil {
		parts = append(parts, r.err.Error())
		partsFiltered = append(partsFiltered, r.err.Error())
	}

	return join(parts), join(partsFiltered)
}

func (r *no) Error() string {
	return r.err.Error()
}
