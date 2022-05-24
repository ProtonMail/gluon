package remote

import (
	"encoding/gob"
)

type OpMailboxCreate struct {
	TempID string
	Name   []string
}

func init() {
	gob.Register(&OpMailboxCreate{})
}

func (op *OpMailboxCreate) merge(other operation) (operation, bool) {
	return nil, false
}

func (OpMailboxCreate) _isOperation() {}
