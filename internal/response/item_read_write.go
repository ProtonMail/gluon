package response

type itemReadWrite struct{}

func ItemReadWrite() *itemReadWrite {
	return &itemReadWrite{}
}

func (c *itemReadWrite) String() string {
	return "READ-WRITE"
}
