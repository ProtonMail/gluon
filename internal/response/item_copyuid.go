package response

import (
	"fmt"

	"github.com/ProtonMail/gluon/imap"
)

type itemCopyUID struct {
	uidValidity        imap.UID
	sourceSet, destSet imap.SeqSet
}

func ItemCopyUID(uidValidity imap.UID, sourceSet, destSet []imap.UID) *itemCopyUID {
	return &itemCopyUID{
		uidValidity: uidValidity,
		sourceSet:   imap.NewSeqSetFromUID(sourceSet),
		destSet:     imap.NewSeqSetFromUID(destSet),
	}
}

func (c *itemCopyUID) String(_ bool) string {
	return fmt.Sprintf("COPYUID %v %v %v", c.uidValidity, c.sourceSet, c.destSet)
}
