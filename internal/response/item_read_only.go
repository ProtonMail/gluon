package response

type itemReadOnly struct{}

func ItemReadOnly() *itemReadOnly {
	return &itemReadOnly{}
}

func (c *itemReadOnly) String(_ bool) string {
	return "READ-ONLY"
}
