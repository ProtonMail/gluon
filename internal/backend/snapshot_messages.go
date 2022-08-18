package backend

import (
	"sync"

	"github.com/ProtonMail/gluon/imap"
	"github.com/bradenaw/juniper/xslices"
)

// snapMsg is a single message inside a snapshot.
type snapMsg struct {
	ID    MessageIDPair
	UID   int
	Seq   int
	flags imap.FlagSet
}

// snapMsgList is an ordered list of messages inside a snapshot.
type snapMsgList struct {
	msg  []*snapMsg
	idx  map[imap.InternalMessageID]int
	lock sync.RWMutex
}

func newMsgList() *snapMsgList {
	return &snapMsgList{
		idx: make(map[imap.InternalMessageID]int),
	}
}

func (list *snapMsgList) insert(msgID MessageIDPair, msgUID int, flags imap.FlagSet) {
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

	list.idx[msgID.InternalID] = len(list.idx)
}

func (list *snapMsgList) remove(msgID imap.InternalMessageID) bool {
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

			if list.idx[message.ID.InternalID] -= 1; list.idx[message.ID.InternalID] < 0 {
				panic("index must be non-negative")
			}
		}
	}

	return true
}

func (list *snapMsgList) update(internalID imap.InternalMessageID, remoteID imap.MessageID) bool {
	list.lock.Lock()
	defer list.lock.Unlock()

	idx, ok := list.idx[internalID]
	if !ok {
		return false
	}

	list.msg[idx].ID.RemoteID = remoteID

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

func (list *snapMsgList) has(msgID imap.InternalMessageID) bool {
	list.lock.RLock()
	defer list.lock.RUnlock()

	_, ok := list.idx[msgID]

	return ok
}

func (list *snapMsgList) get(msgID imap.InternalMessageID) (*snapMsg, bool) {
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
