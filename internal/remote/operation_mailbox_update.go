package remote

import "encoding/gob"

type OpMailboxUpdate struct {
	OperationBase
	MBoxID string
	Name   []string
}

func init() {
	gob.Register(&OpMailboxUpdate{})
}

func (op *OpMailboxUpdate) merge(other operation) (operation, bool) {
	return nil, false
}

func (OpMailboxUpdate) _isOperation() {}
