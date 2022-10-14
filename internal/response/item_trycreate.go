package response

type itemTryCreate struct{}

func ItemTryCreate() *itemTryCreate {
	return &itemTryCreate{}
}

func (c *itemTryCreate) String(_ bool) string {
	return "TRYCREATE"
}
