package response

type itemBadCharset struct{}

func ItemBadCharset() *itemBadCharset {
	return &itemBadCharset{}
}

func (c *itemBadCharset) String() string {
	return "BADCHARSET"
}
