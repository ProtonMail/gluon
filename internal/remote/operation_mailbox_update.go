package remote

import "encoding/gob"

type OpMailboxUpdate struct {
	MBoxID string
	Name   []string
}

func init() {
	gob.Register(&OpMailboxUpdate{})
}

func (op *OpMailboxUpdate) merge(other operation) (operation, bool) {
	return nil, false
}

func (op *OpMailboxUpdate) setMailboxID(tempID, mailboxID string) {
	if op.MBoxID == tempID {
		op.MBoxID = mailboxID
	}
}

func (OpMailboxUpdate) _isOperation() {}
