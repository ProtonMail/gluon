package response

type itemReadOnly struct{}

func ItemReadOnly() *itemReadOnly {
	return &itemReadOnly{}
}

func (c *itemReadOnly) String() string {
	return "READ-ONLY"
}
