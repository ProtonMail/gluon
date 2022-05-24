package response

import (
	"fmt"

	"github.com/ProtonMail/gluon/imap"
)

type itemCopyUID struct {
	uidValidity        int
	sourceSet, destSet imap.SeqSet
}

func ItemCopyUID(uidValidity int, sourceSet, destSet []int) *itemCopyUID {
	return &itemCopyUID{
		uidValidity: uidValidity,
		sourceSet:   imap.NewSeqSet(sourceSet),
		destSet:     imap.NewSeqSet(destSet),
	}
}

func (c *itemCopyUID) String() string {
	return fmt.Sprintf("COPYUID %v %v %v", c.uidValidity, c.sourceSet, c.destSet)
}
