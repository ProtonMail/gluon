package remote

import (
	"encoding/gob"
	"github.com/ProtonMail/gluon/imap"
)

type OpMailboxDelete struct {
	OperationBase
	MBoxID imap.LabelID
}

func init() {
	gob.Register(&OpMailboxDelete{})
}

func (op *OpMailboxDelete) merge(other operation) (operation, bool) {
	return nil, false
}

func (OpMailboxDelete) _isOperation() {}
