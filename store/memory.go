package store

import (
	"errors"
	"sync"
)

type inMemoryStore struct {
	data map[string][]byte
	lock sync.RWMutex
}

func NewInMemoryStore() Store {
	return &inMemoryStore{
		data: make(map[string][]byte),
	}
}

func (c *inMemoryStore) Get(messageID string) ([]byte, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	literal, ok := c.data[messageID]
	if !ok {
		return nil, errors.New("no such message in cache")
	}

	return literal, nil
}

func (c *inMemoryStore) Set(messageID string, literal []byte) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.data[messageID] = literal

	return nil
}

func (c *inMemoryStore) Update(oldID, newID string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	message, ok := c.data[oldID]
	if !ok {
		panic("no such message")
	}

	c.data[newID] = message

	delete(c.data, oldID)

	return nil
}

func (c *inMemoryStore) Delete(ids ...string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	for _, id := range ids {
		delete(c.data, id)
	}

	return nil
}
