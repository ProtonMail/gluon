package response

type itemReadWrite struct{}

func ItemReadWrite() *itemReadWrite {
	return &itemReadWrite{}
}

func (c *itemReadWrite) Strings() (raw string, _ string) {
	raw = "READ-WRITE"
	return raw, raw
}
