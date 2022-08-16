package remote

import (
	"encoding/gob"
	"github.com/ProtonMail/gluon/imap"
)

type OpMessageFlagged struct {
	OperationBase
	MessageIDs []imap.MessageID
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

func (OpMessageFlagged) _isOperation() {}
