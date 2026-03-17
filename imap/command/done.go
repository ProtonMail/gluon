package command

type Done struct{}

func (l Done) String() string {
	return "DONE"
}

func (l Done) SanitizedString() string {
	return l.String()
}
