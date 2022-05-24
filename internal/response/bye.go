package response

type bye struct {
	msg string
}

func Bye() *bye {
	return &bye{}
}

func (r *bye) WithMessage(msg string) *bye {
	r.msg = msg
	return r
}

func (r *bye) Send(s Session) error {
	return s.WriteResponse(r.String())
}

func (r *bye) String() string {
	parts := []string{"*", "BYE", faceBye}

	if r.msg != "" {
		parts = append(parts, r.msg)
	}

	return join(parts)
}
