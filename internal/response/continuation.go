package response

import "strings"

type continuation struct {
	tag string
}

func Continuation() *continuation {
	return &continuation{
		tag: "+",
	}
}

func (r *continuation) Send(s Session) error {
	return s.WriteResponse(r)
}

func (r *continuation) Strings() (raw string, _ string) {
	raw = strings.Join([]string{r.tag, "Ready"}, " ")
	return raw, raw
}
