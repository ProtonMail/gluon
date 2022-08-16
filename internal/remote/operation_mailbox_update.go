package remote

import (
	"encoding/gob"
	"github.com/ProtonMail/gluon/imap"
)

type OpMailboxUpdate struct {
	OperationBase
	MBoxID imap.LabelID
	Name   []string
}

func init() {
	gob.Register(&OpMailboxUpdate{})
}

func (op *OpMailboxUpdate) merge(other operation) (operation, bool) {
	return nil, false
}

func (OpMailboxUpdate) _isOperation() {}
