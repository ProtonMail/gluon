package remote

import "encoding/gob"

type OpConnMetadataStoreDelete struct {
	OperationBase
}

func (op *OpConnMetadataStoreDelete) merge(operation) (operation, bool) {
	return nil, false
}

func (OpConnMetadataStoreDelete) _isOperation() {}

type OpConnMetadataStoreCreate struct {
	OperationBase
}

func (op *OpConnMetadataStoreCreate) merge(operation) (operation, bool) {
	return nil, false
}

func (OpConnMetadataStoreCreate) _isOperation() {}

type OpConnMetadataStoreSetValue struct {
	OperationBase
	Key   string
	Value any
}

func (op *OpConnMetadataStoreSetValue) merge(operation) (operation, bool) {
	return nil, false
}

func (OpConnMetadataStoreSetValue) _isOperation() {}

func init() {
	gob.Register(&OpConnMetadataStoreDelete{})
	gob.Register(&OpConnMetadataStoreSetValue{})
	gob.Register(&OpConnMetadataStoreCreate{})
}
