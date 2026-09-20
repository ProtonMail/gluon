package response

type itemOverQuota struct{}

func ItemOverQuota() *itemOverQuota {
	return &itemOverQuota{}
}

func (c *itemOverQuota) String() string {
	return "OVERQUOTA"
}
