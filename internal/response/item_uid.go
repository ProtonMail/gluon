package response

import "fmt"

type itemUID struct {
	uid int
}

func ItemUID(n int) *itemUID {
	return &itemUID{uid: n}
}

func (c *itemUID) String() string {
	return fmt.Sprintf("UID %v", c.uid)
}
