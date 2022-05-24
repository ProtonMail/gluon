package response

import "fmt"

type itemUIDNext struct {
	uid int
}

func ItemUIDNext(n int) *itemUIDNext {
	return &itemUIDNext{uid: n}
}

func (c *itemUIDNext) String() string {
	return fmt.Sprintf("UIDNEXT %v", c.uid)
}
