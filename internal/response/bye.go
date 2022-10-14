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
	return s.WriteResponse(r)
}

func (r *bye) Strings() (raw string, _ string) {
	parts := []string{"*", "BYE"}

	if r.msg != "" {
		parts = append(parts, r.msg)
	}

	raw = join(parts)

	return raw, raw
}

func (r *bye) WithMailboxDeleted() *bye {
	r.msg = "Mailbox was deleted, have to disconnect."

	return r
}

func (r *bye) WithInconsistentState() *bye {
	r.msg = "IMAP session state is inconsistent, please re-login."

	return r
}
