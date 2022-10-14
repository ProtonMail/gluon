package response

type itemReadWrite struct{}

func ItemReadWrite() *itemReadWrite {
	return &itemReadWrite{}
}

func (c *itemReadWrite) String(_ bool) string {
	return "READ-WRITE"
}
