package response

import "fmt"

type itemUnseen struct {
	count uint32
}

func ItemUnseen(n uint32) *itemUnseen {
	return &itemUnseen{count: n}
}

func (c *itemUnseen) String() string {
	return fmt.Sprintf("UNSEEN %v", c.count)
}
