package remote

import "encoding/gob"

type OpMessageAdd struct {
	MessageIDs []string
	MBoxID     string
}

func init() {
	gob.Register(&OpMessageAdd{})
}

func (op *OpMessageAdd) merge(other operation) (operation, bool) {
	switch other := other.(type) {
	case *OpMessageAdd:
		if op.MBoxID != other.MBoxID {
			return nil, false
		}

		op.MessageIDs = append(op.MessageIDs, other.MessageIDs...)

		return op, true

	default:
		return nil, false
	}
}

func (op *OpMessageAdd) setMailboxID(tempID, mboxID string) {
	if op.MBoxID == tempID {
		op.MBoxID = mboxID
	}
}

func (op *OpMessageAdd) setMessageID(tempID, messageID string) {
	for idx := range op.MessageIDs {
		if op.MessageIDs[idx] == tempID {
			op.MessageIDs[idx] = messageID
		}
	}
}

func (OpMessageAdd) _isOperation() {}
