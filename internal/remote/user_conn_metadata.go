package remote

import "fmt"

func (user *User) CreateConnMetadataStore(metadataID ConnMetadataID) error {
	user.connMetadataStoreLock.Lock()
	defer user.connMetadataStoreLock.Unlock()

	user.connMetadataStore.CreateStore(metadataID)

	return nil
}

func (user *User) DeleteConnMetadataStore(metadataID ConnMetadataID) error {
	user.connMetadataStoreLock.Lock()
	defer user.connMetadataStoreLock.Unlock()

	user.connMetadataStore.DeleteStore(metadataID)

	return nil
}

func (user *User) SetConnMetadataValue(metadataID ConnMetadataID, key string, value any) error {
	user.connMetadataStoreLock.Lock()
	defer user.connMetadataStoreLock.Unlock()

	if ok := user.connMetadataStore.SetValue(metadataID, key, value); !ok {
		return fmt.Errorf("failed to set value for ConnMetadata with ID=%v", metadataID)
	}

	return nil
}
