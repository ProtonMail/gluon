package response

import (
	"fmt"

	"github.com/ProtonMail/gluon/imap"
)

type itemAppendUID struct {
	uidValidity, messageUID imap.UID
}

func ItemAppendUID(uidValidity, messageUID imap.UID) *itemAppendUID {
	return &itemAppendUID{
		uidValidity: uidValidity,
		messageUID:  messageUID,
	}
}

func (c *itemAppendUID) String(_ bool) string {
	return fmt.Sprintf("APPENDUID %v %v", c.uidValidity, c.messageUID)
}
