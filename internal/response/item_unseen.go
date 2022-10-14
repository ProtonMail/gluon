package response

import "fmt"

type itemUnseen struct {
	count uint32
}

func ItemUnseen(n uint32) *itemUnseen {
	return &itemUnseen{count: n}
}

func (c *itemUnseen) Strings() (raw string, _ string) {
	raw = fmt.Sprintf("UNSEEN %v", c.count)
	return raw, raw
}
