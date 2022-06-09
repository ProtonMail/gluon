package remote

import "encoding/gob"

type OpMessageFlagged struct {
	OperationBase
	MessageIDs []string
	Flagged    bool
}

func init() {
	gob.Register(&OpMessageFlagged{})
}

func (op *OpMessageFlagged) merge(other operation) (operation, bool) {
	switch other := other.(type) {
	case *OpMessageFlagged:
		if op.Flagged != other.Flagged || op.MetadataID != other.MetadataID {
			return nil, false
		}

		op.MessageIDs = append(op.MessageIDs, other.MessageIDs...)

		return op, true

	default:
		return nil, false
	}
}

func (op *OpMessageFlagged) setMessageID(tempID, messageID string) {
	for idx := range op.MessageIDs {
		if op.MessageIDs[idx] == tempID {
			op.MessageIDs[idx] = messageID
		}
	}
}

func (OpMessageFlagged) _isOperation() {}
