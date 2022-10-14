package response

type itemTryCreate struct{}

func ItemTryCreate() *itemTryCreate {
	return &itemTryCreate{}
}

func (c *itemTryCreate) Strings() (raw string, _ string) {
	raw = "TRYCREATE"
	return raw, raw
}
