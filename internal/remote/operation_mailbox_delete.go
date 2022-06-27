package remote

import "encoding/gob"

type OpMailboxDelete struct {
	OperationBase
	MBoxID string
}

func init() {
	gob.Register(&OpMailboxDelete{})
}

func (op *OpMailboxDelete) merge(other operation) (operation, bool) {
	return nil, false
}

func (OpMailboxDelete) _isOperation() {}
