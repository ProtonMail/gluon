package remote

import "encoding/gob"

type OpRemMailboxTempID struct {
	OperationBase
	tempID string
}

type OpRemMessageTempID struct {
	OperationBase
	tempID string
}

func (op *OpRemMailboxTempID) merge(other operation) (operation, bool) {
	switch other := other.(type) {
	case *OpRemMailboxTempID:
		if op.tempID != other.tempID || op.MetadataID != other.MetadataID {
			return nil, false
		}

		return op, true
	default:
		return nil, false
	}
}

func (OpRemMailboxTempID) _isOperation() {}

func (op *OpRemMessageTempID) merge(other operation) (operation, bool) {
	switch other := other.(type) {
	case *OpRemMessageTempID:
		if op.tempID != other.tempID || op.MetadataID != other.MetadataID {
			return nil, false
		}

		return op, true
	default:
		return nil, false
	}
}

func (OpRemMessageTempID) _isOperation() {}

func init() {
	gob.Register(&OpRemMailboxTempID{})
	gob.Register(&OpRemMessageTempID{})
}
