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

func (c *itemUIDValidity) Strings() (raw string, _ string) {
	raw = fmt.Sprintf("UIDVALIDITY %v", c.val)
	return raw, raw
}
