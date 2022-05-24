package backend

import (
	"sync"

	"github.com/ProtonMail/gluon/imap"
	"github.com/bradenaw/juniper/xslices"
)

// snapMsg is a single message inside a snapshot.
type snapMsg struct {
	ID    string
	UID   int
	Seq   int
	flags imap.FlagSet
}

// snapMsgList is an ordered list of messages inside a snapshot.
type snapMsgList struct {
	msg  []*snapMsg
	idx  map[string]int
	lock sync.RWMutex
}

func newMsgList() *snapMsgList {
	return &snapMsgList{
		idx: make(map[string]int),
	}
}

func (list *snapMsgList) insert(msgID string, msgUID int, flags imap.FlagSet) {
	list.lock.Lock()
	defer list.lock.Unlock()

	if len(list.msg) > 0 && list.msg[len(list.msg)-1].UID >= msgUID {
		panic("UIDs must be strictly ascending")
	}

	list.msg = append(list.msg, &snapMsg{
		ID:    msgID,
		UID:   msgUID,
		Seq:   len(list.msg) + 1,
		flags: flags,
	})

	list.idx[msgID] = len(list.idx)
}

func (list *snapMsgList) remove(msgID string) bool {
	list.lock.Lock()
	defer list.lock.Unlock()

	idx, ok := list.idx[msgID]
	if !ok {
		return false
	}

	delete(list.idx, msgID)

	list.msg = append(
		list.msg[:idx],
		list.msg[idx+1:]...,
	)

	if len(list.msg) > 0 {
		for _, message := range list.msg[idx:] {
			if message.Seq -= 1; message.Seq < 1 {
				panic("sequence number must be positive")
			}

			if list.idx[message.ID] -= 1; list.idx[message.ID] < 0 {
				panic("index must be non-negative")
			}
		}
	}

	return true
}

func (list *snapMsgList) update(oldID, newID string) bool {
	list.lock.Lock()
	defer list.lock.Unlock()

	idx, ok := list.idx[oldID]
	if !ok {
		return false
	}

	list.msg[idx].ID = newID

	list.idx[newID] = idx

	delete(list.idx, oldID)

	return true
}

func (list *snapMsgList) all() []*snapMsg {
	list.lock.RLock()
	defer list.lock.RUnlock()

	return list.msg
}

func (list *snapMsgList) len() int {
	list.lock.RLock()
	defer list.lock.RUnlock()

	return len(list.msg)
}

func (list *snapMsgList) where(fn func(*snapMsg) bool) []*snapMsg {
	list.lock.RLock()
	defer list.lock.RUnlock()

	return xslices.Filter(list.msg, fn)
}

func (list *snapMsgList) has(msgID string) bool {
	list.lock.RLock()
	defer list.lock.RUnlock()

	_, ok := list.idx[msgID]

	return ok
}

func (list *snapMsgList) get(msgID string) (*snapMsg, bool) {
	list.lock.RLock()
	defer list.lock.RUnlock()

	idx, ok := list.idx[msgID]
	if !ok {
		return nil, false
	}

	return list.msg[idx], true
}

func (list *snapMsgList) seq(seq int) (*snapMsg, bool) {
	list.lock.RLock()
	defer list.lock.RUnlock()

	if len(list.msg) < seq {
		return nil, false
	}

	return list.msg[seq-1], true
}

func (list *snapMsgList) last() *snapMsg {
	list.lock.RLock()
	defer list.lock.RUnlock()

	return list.msg[len(list.msg)-1]
}
