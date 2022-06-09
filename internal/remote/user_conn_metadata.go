package remote

func (user *User) CreateConnMetadataStore(metadataID ConnMetadataID) error {
	if err := user.pushOp(&OpConnMetadataStoreCreate{
		OperationBase: OperationBase{MetadataID: metadataID},
	}); err != nil {
		return err
	}

	return nil
}

func (user *User) DeleteConnMetadataStore(metadataID ConnMetadataID) error {
	return user.pushOp(&OpConnMetadataStoreDelete{
		OperationBase: OperationBase{MetadataID: metadataID},
	})
}

func (user *User) SetConnMetadataValue(metadataID ConnMetadataID, key string, value any) error {
	return user.pushOp(&OpConnMetadataStoreSetValue{
		OperationBase: OperationBase{MetadataID: metadataID},
		Key:           key,
		Value:         value,
	})
}
