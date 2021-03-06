package remote

import "encoding/gob"

type OpMessageAdd struct {
	OperationBase
	MessageIDs []string
	MBoxID     string
}

func init() {
	gob.Register(&OpMessageAdd{})
}

func (op *OpMessageAdd) merge(other operation) (operation, bool) {
	switch other := other.(type) {
	case *OpMessageAdd:
		if op.MBoxID != other.MBoxID || op.MetadataID != other.MetadataID {
			return nil, false
		}

		op.MessageIDs = append(op.MessageIDs, other.MessageIDs...)

		return op, true

	default:
		return nil, false
	}
}

func (OpMessageAdd) _isOperation() {}
