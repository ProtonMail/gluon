package store

import (
	"errors"
	"sync"

	"github.com/ProtonMail/gluon/imap"
)

type inMemoryStore struct {
	data map[imap.InternalMessageID][]byte
	lock sync.RWMutex
}

func NewInMemoryStore() Store {
	return &inMemoryStore{
		data: make(map[imap.InternalMessageID][]byte),
	}
}

func (c *inMemoryStore) Get(messageID imap.InternalMessageID) ([]byte, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	literal, ok := c.data[messageID]
	if !ok {
		return nil, errors.New("no such message in cache")
	}

	return literal, nil
}

func (c *inMemoryStore) Set(messageID imap.InternalMessageID, literal []byte) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.data[messageID] = literal

	return nil
}

func (c *inMemoryStore) Delete(ids ...imap.InternalMessageID) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	for _, id := range ids {
		delete(c.data, id)
	}

	return nil
}

func (c *inMemoryStore) Close() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.data = make(map[imap.InternalMessageID][]byte)

	return nil
}
