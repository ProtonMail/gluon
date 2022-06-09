package remote

import (
	"encoding/gob"
	"time"

	"github.com/ProtonMail/gluon/imap"
)

type OpMessageCreate struct {
	OperationBase
	TempID  string
	MBoxID  string
	Literal []byte
	Flags   imap.FlagSet
	Date    time.Time
}

func init() {
	gob.Register(&OpMessageCreate{})
}

func (op *OpMessageCreate) merge(other operation) (operation, bool) {
	return nil, false
}

func (op *OpMessageCreate) setMailboxID(tempID, mailboxID string) {
	if op.MBoxID == tempID {
		op.MBoxID = mailboxID
	}
}

func (OpMessageCreate) _isOperation() {}
