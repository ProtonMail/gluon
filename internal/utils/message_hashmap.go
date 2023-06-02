package utils

import (
	"sync"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/rfc822"
)

// MessageHashesMap tracks the hashes for a literal and it's associated internal IMAP ID.
type MessageHashesMap struct {
	lock     sync.Mutex
	idToHash map[imap.InternalMessageID]string
	hashes   map[string]struct{}
}

func NewMessageHashesMap() *MessageHashesMap {
	return &MessageHashesMap{
		idToHash: make(map[imap.InternalMessageID]string),
		hashes:   make(map[string]struct{}),
	}
}

// Insert inserts the hash of the current message literal into the map and return true if an existing value was already
// present.
func (m *MessageHashesMap) Insert(id imap.InternalMessageID, literal []byte) (bool, error) {
	literalHashStr, err := rfc822.GetMessageHash(literal)
	if err != nil {
		return false, err
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.hashes[literalHashStr]; ok {
		return true, nil
	}

	m.idToHash[id] = literalHashStr
	m.hashes[literalHashStr] = struct{}{}

	return false, nil
}

// Erase removes the info associated with a given id.
func (m *MessageHashesMap) Erase(ids ...imap.InternalMessageID) {
	m.lock.Lock()
	defer m.lock.Unlock()

	for _, id := range ids {
		if v, ok := m.idToHash[id]; ok {
			delete(m.hashes, v)
		}

		delete(m.idToHash, id)
	}
}
