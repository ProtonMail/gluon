package remote

import (
	"bytes"
	"encoding/gob"
)

type ConnMetadataID uint32

// connMetadataStore provides a storage container for any type of data that needs to be associated to a connection.
type connMetadataStore struct {
	data map[ConnMetadataID]map[string]any
}

func newConnMetadataStore() connMetadataStore {
	return connMetadataStore{
		data: make(map[ConnMetadataID]map[string]any),
	}
}

func (c *connMetadataStore) CreateStore(id ConnMetadataID) {
	c.data[id] = make(map[string]any)
}

func (c *connMetadataStore) DeleteStore(id ConnMetadataID) {
	delete(c.data, id)
}

func (c *connMetadataStore) GetActiveStoreIDs() []ConnMetadataID {
	var values []ConnMetadataID

	for k, _ := range c.data {
		values = append(values, k)
	}

	return values
}

func (c *connMetadataStore) SetValue(id ConnMetadataID, key string, value any) bool {
	valueStore, ok := c.data[id]

	if !ok {
		return false
	}

	valueStore[key] = value

	return true
}

func (c *connMetadataStore) GetValue(id ConnMetadataID, key string) any {
	valueStore, ok := c.data[id]

	if !ok {
		return false
	}

	value, ok := valueStore[key]

	if !ok {
		return nil
	}

	return value
}

func (c *connMetadataStore) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	if err := gob.NewEncoder(buf).Encode(c.data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (c *connMetadataStore) UnmarshalBinary(data []byte) error {
	return gob.NewDecoder(bytes.NewReader(data)).Decode(&c.data)
}

func init() {
	gob.Register(&connMetadataStore{})
}
