package response

type bad struct {
	tag string
	err error
}

func Bad(withTag ...string) *bad {
	var tag string

	if len(withTag) > 0 {
		tag = withTag[0]
	} else {
		tag = "*"
	}

	return &bad{
		tag: tag,
	}
}

func (r *bad) WithError(err error) *bad {
	r.err = err
	return r
}

func (r *bad) Send(s Session) error {
	return s.WriteResponse(r)
}

func (r *bad) String(_ bool) string {
	parts := []string{r.tag, "BAD"}

	if r.err != nil {
		parts = append(parts, r.err.Error())
	}

	return join(parts)
}

func (r *bad) Error() string {
	return r.err.Error()
}
