package remote

import "encoding/gob"

type OpMessageRemove struct {
	OperationBase
	MessageIDs []string
	MBoxID     string
}

func init() {
	gob.Register(&OpMessageRemove{})
}

func (op *OpMessageRemove) merge(other operation) (operation, bool) {
	switch other := other.(type) {
	case *OpMessageRemove:
		if op.MBoxID != other.MBoxID || op.MetadataID != other.MetadataID {
			return nil, false
		}

		op.MessageIDs = append(op.MessageIDs, other.MessageIDs...)

		return op, true

	default:
		return nil, false
	}
}

func (op *OpMessageRemove) getConnMetadataID() ConnMetadataID {
	return op.MetadataID
}

func (OpMessageRemove) _isOperation() {}
