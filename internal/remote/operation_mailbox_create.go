package remote

import (
	"encoding/gob"

	"github.com/ProtonMail/gluon/imap"
)

type OpMailboxCreate struct {
	OperationBase
	InternalID imap.InternalMailboxID
	Name       []string
}

func init() {
	gob.Register(&OpMailboxCreate{})
}

func (op *OpMailboxCreate) merge(other operation) (operation, bool) {
	return nil, false
}

func (OpMailboxCreate) _isOperation() {}
