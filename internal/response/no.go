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

func (r *no) String(isPrivateByDefault bool) (res string) {
	parts := []string{r.tag, "NO"}

	if len(r.items) > 0 {
		var items []string

		for _, item := range r.items {
			items = append(items, item.String(isPrivateByDefault))
		}

		parts = append(parts, fmt.Sprintf("[%v]", join(items)))
	}

	if r.err != nil {
		parts = append(parts, r.err.Error())
	}

	return join(parts)
}

func (r *no) Error() string {
	return r.err.Error()
}
