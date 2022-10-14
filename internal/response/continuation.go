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

func (r *continuation) String(_ bool) string {
	return strings.Join([]string{r.tag, "Ready"}, " ")
}
