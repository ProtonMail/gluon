package response

type continuation struct {
	tag string
}

func Continuation() *continuation {
	return &continuation{
		tag: "+",
	}
}

func (r *continuation) Send(s Session) error {
	return s.WriteResponse(r.String())
}

func (r *continuation) String() string {
	return join([]string{r.tag, faceCon})
}
