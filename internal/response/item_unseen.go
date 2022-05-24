package response

import "fmt"

type itemUnseen struct {
	count int
}

func ItemUnseen(n int) *itemUnseen {
	return &itemUnseen{count: n}
}

func (c *itemUnseen) String() string {
	return fmt.Sprintf("UNSEEN %v", c.count)
}
