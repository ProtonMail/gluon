package response

import (
	"fmt"

	"github.com/ProtonMail/gluon/imap"
)

type itemUID struct {
	uid imap.UID
}

func ItemUID(n imap.UID) *itemUID {
	return &itemUID{uid: n}
}

func (c *itemUID) Strings() (raw string, _ string) {
	raw = fmt.Sprintf("UID %v", c.uid)
	return raw, raw
}
