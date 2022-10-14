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

func (r *ok) String(isPrivateByDefault bool) string {
	parts := []string{r.tag, "OK"}

	if len(r.items) > 0 {
		var items []string

		for _, item := range r.items {
			items = append(items, item.String(isPrivateByDefault))
		}

		parts = append(parts, fmt.Sprintf("[%v]", join(items)))
	}

	if r.msg != "" {
		parts = append(parts, r.msg)
	}

	return join(parts)
}
