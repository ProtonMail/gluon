package response

type itemExpungeIssued struct{}

func ItemExpungeIssued() *itemExpungeIssued {
	return &itemExpungeIssued{}
}

func (c *itemExpungeIssued) String() string {
	return "EXPUNGEISSUED"
}
