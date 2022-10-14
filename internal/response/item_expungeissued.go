package response

type itemExpungeIssued struct{}

func ItemExpungeIssued() *itemExpungeIssued {
	return &itemExpungeIssued{}
}

func (c *itemExpungeIssued) String(_ bool) string {
	return "EXPUNGEISSUED"
}
