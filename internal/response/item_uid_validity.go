package response

import (
	"fmt"

	"github.com/ProtonMail/gluon/imap"
)

type itemUIDValidity struct {
	val imap.UID
}

func ItemUIDValidity(n imap.UID) *itemUIDValidity {
	return &itemUIDValidity{val: n}
}

func (c *itemUIDValidity) String() string {
	return fmt.Sprintf("UIDVALIDITY %v", c.val)
}
