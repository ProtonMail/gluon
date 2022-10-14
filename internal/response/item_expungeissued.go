package response

type itemExpungeIssued struct{}

func ItemExpungeIssued() *itemExpungeIssued {
	return &itemExpungeIssued{}
}

func (c *itemExpungeIssued) Strings() (raw string, _ string) {
	raw = "EXPUNGEISSUED"
	return raw, raw
}
