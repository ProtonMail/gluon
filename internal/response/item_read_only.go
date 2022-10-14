package response

type itemReadOnly struct{}

func ItemReadOnly() *itemReadOnly {
	return &itemReadOnly{}
}

func (c *itemReadOnly) Strings() (raw string, _ string) {
	raw = "READ-ONLY"
	return raw, raw
}
