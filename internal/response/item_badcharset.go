package response

type itemBadCharset struct{}

func ItemBadCharset() *itemBadCharset {
	return &itemBadCharset{}
}

func (c *itemBadCharset) Strings() (raw string, _ string) {
	raw = "BADCHARSET"
	return raw, raw
}
