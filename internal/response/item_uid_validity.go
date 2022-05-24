package response

import "fmt"

type itemUIDValidity struct {
	val int
}

func ItemUIDValidity(n int) *itemUIDValidity {
	return &itemUIDValidity{val: n}
}

func (c *itemUIDValidity) String() string {
	return fmt.Sprintf("UIDVALIDITY %v", c.val)
}
