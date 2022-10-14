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

func (c *itemCopyUID) Strings() (raw string, _ string) {
	raw = fmt.Sprintf("COPYUID %v %v %v", c.uidValidity, c.sourceSet, c.destSet)
	return raw, raw
}
