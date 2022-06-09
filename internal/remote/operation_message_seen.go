package remote

import "encoding/gob"

type OpMessageSeen struct {
	OperationBase
	MessageIDs []string
	Seen       bool
}

func init() {
	gob.Register(&OpMessageSeen{})
}

func (op *OpMessageSeen) merge(other operation) (operation, bool) {
	switch other := other.(type) {
	case *OpMessageSeen:
		if op.Seen != other.Seen || op.MetadataID != other.MetadataID {
			return nil, false
		}

		op.MessageIDs = append(op.MessageIDs, other.MessageIDs...)

		return op, true

	default:
		return nil, false
	}
}

func (op *OpMessageSeen) setMessageID(tempID, messageID string) {
	for idx := range op.MessageIDs {
		if op.MessageIDs[idx] == tempID {
			op.MessageIDs[idx] = messageID
		}
	}
}

func (OpMessageSeen) _isOperation() {}
