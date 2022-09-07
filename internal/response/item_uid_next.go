package response

import (
	"fmt"
	"github.com/ProtonMail/gluon/imap"
)

type itemUIDNext struct {
	uid imap.UID
}

func ItemUIDNext(n imap.UID) *itemUIDNext {
	return &itemUIDNext{uid: n}
}

func (c *itemUIDNext) String() string {
	return fmt.Sprintf("UIDNEXT %v", c.uid)
}
