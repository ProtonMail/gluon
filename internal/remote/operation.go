package remote

import (
	"bytes"
	"encoding/gob"
)

type operation interface {
	merge(op operation) (operation, bool)

	getConnMetadataID() ConnMetadataID

	_isOperation()
}

type OperationBase struct {
	MetadataID ConnMetadataID
}

func (c *OperationBase) getConnMetadataID() ConnMetadataID {
	return c.MetadataID
}

type mailboxOperation interface {
	setMailboxID(tempID, mailboxID string)
}

type messageOperation interface {
	setMessageID(tempID, messageID string)
}

// saveOps serializes the operation queue to a binary format to be serialized to disk.
func saveOps(ops []operation) ([]byte, error) {
	buf := new(bytes.Buffer)

	if err := gob.NewEncoder(buf).Encode(ops); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// loadOps deserializes the operation queue to a binary format to be serialized to disk.
func loadOps(b []byte) ([]operation, error) {
	var ops []operation

	if err := gob.NewDecoder(bytes.NewReader(b)).Decode(&ops); err != nil {
		return nil, err
	}

	return ops, nil
}
