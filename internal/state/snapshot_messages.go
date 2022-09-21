package state

import (
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/ids"
	"github.com/bradenaw/juniper/xslices"
	"golang.org/x/exp/slices"
)

// snapMsg is a single message inside a snapshot.
type snapMsg struct {
	ID    ids.MessageIDPair
	UID   imap.UID
	Seq   imap.SeqID
	flags imap.FlagSet
}

// snapMsgList is an ordered list of messages inside a snapshot.
type snapMsgList struct {
	msg      []*snapMsg
	idx      map[imap.InternalMessageID]int
	uidToMsg map[imap.UID]*snapMsg
}

func newMsgList(capacity int) *snapMsgList {
	return &snapMsgList{
		idx:      make(map[imap.InternalMessageID]int, capacity),
		msg:      make([]*snapMsg, 0, capacity),
		uidToMsg: make(map[imap.UID]*snapMsg, capacity),
	}
}

func (list *snapMsgList) insert(msgID ids.MessageIDPair, msgUID imap.UID, flags imap.FlagSet) {
	if len(list.msg) > 0 && list.msg[len(list.msg)-1].UID >= msgUID {
		panic("UIDs must be strictly ascending")
	}

	snapMsg := &snapMsg{
		ID:    msgID,
		UID:   msgUID,
		Seq:   imap.SeqID(len(list.msg) + 1),
		flags: flags,
	}

	list.msg = append(list.msg, snapMsg)

	list.idx[msgID.InternalID] = len(list.idx)
	list.uidToMsg[msgUID] = snapMsg
}

func (list *snapMsgList) remove(msgID imap.InternalMessageID) bool {
	idx, ok := list.idx[msgID]
	if !ok {
		return false
	}

	delete(list.uidToMsg, list.msg[idx].UID)
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
	idx, ok := list.idx[internalID]
	if !ok {
		return false
	}

	list.msg[idx].ID.RemoteID = remoteID

	return true
}

func (list *snapMsgList) all() []*snapMsg {
	return list.msg
}

func (list *snapMsgList) len() int {
	return len(list.msg)
}

func (list *snapMsgList) where(fn func(*snapMsg) bool) []*snapMsg {
	return xslices.Filter(list.msg, fn)
}

func (list *snapMsgList) has(msgID imap.InternalMessageID) bool {
	_, ok := list.idx[msgID]

	return ok
}

func (list *snapMsgList) get(msgID imap.InternalMessageID) (*snapMsg, bool) {
	idx, ok := list.idx[msgID]
	if !ok {
		return nil, false
	}

	return list.msg[idx], true
}

func (list *snapMsgList) seq(seq imap.SeqID) (*snapMsg, bool) {
	if imap.SeqID(len(list.msg)) < seq {
		return nil, false
	}

	return list.msg[seq-1], true
}

func (list *snapMsgList) last() *snapMsg {
	return list.msg[len(list.msg)-1]
}

func (list *snapMsgList) seqRange(seqLo, seqHi imap.SeqID) []*snapMsg {
	return list.msg[seqLo-1 : seqHi]
}

func (list *snapMsgList) uidRange(uidLo, uidHi imap.UID) []*snapMsg {
	listLen := len(list.msg)

	cmpFunc := func(s1 *snapMsg, s2 *snapMsg) int {
		return int(s1.UID) - int(s2.UID)
	}

	targetSnapLo := snapMsg{UID: uidLo}

	indexLo, _ := slices.BinarySearchFunc(list.msg, &targetSnapLo, cmpFunc)

	if indexLo >= listLen {
		return nil
	}

	targetSnapHi := snapMsg{UID: uidHi}

	indexHi, ok := slices.BinarySearchFunc(list.msg[indexLo:], &targetSnapHi, cmpFunc)
	if ok {
		indexHi++
	}

	if indexHi >= listLen {
		indexHi = listLen
	}

	return list.msg[indexLo : indexLo+indexHi]
}

func (list *snapMsgList) getWithUID(uid imap.UID) (*snapMsg, bool) {
	msg, ok := list.uidToMsg[uid]

	return msg, ok
}
