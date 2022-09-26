package state

import (
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/ids"
	"golang.org/x/exp/slices"
)

// snapMsg is a single message inside a snapshot.
type snapMsg struct {
	ID    ids.MessageIDPair
	UID   imap.UID
	flags imap.FlagSet
}

type snapMsgWithSeq struct {
	Seq imap.SeqID
	*snapMsg
}

// snapMsgList is an ordered list of messages inside a snapshot.
type snapMsgList struct {
	msg []*snapMsg
	idx map[imap.InternalMessageID]*snapMsg
}

func newMsgList(capacity int) *snapMsgList {
	return &snapMsgList{
		msg: make([]*snapMsg, 0, capacity),
		idx: make(map[imap.InternalMessageID]*snapMsg),
	}
}

func (list *snapMsgList) binarySearchByUID(uid imap.UID) (int, bool) {
	msg := snapMsg{UID: uid}

	index, ok := slices.BinarySearchFunc(list.msg, &msg, func(s1 *snapMsg, s2 *snapMsg) int {
		return int(s1.UID) - int(s2.UID)
	})

	return index, ok
}

func (list *snapMsgList) insert(msgID ids.MessageIDPair, msgUID imap.UID, flags imap.FlagSet) {
	if len(list.msg) > 0 && list.msg[len(list.msg)-1].UID >= msgUID {
		panic("UIDs must be strictly ascending")
	}

	snapMsg := &snapMsg{
		ID:    msgID,
		UID:   msgUID,
		flags: flags,
	}

	list.msg = append(list.msg, snapMsg)

	list.idx[msgID.InternalID] = snapMsg
}

func (list *snapMsgList) remove(msgID imap.InternalMessageID) bool {
	snapshotMsg, ok := list.idx[msgID]
	if !ok {
		return false
	}

	index, ok := list.binarySearchByUID(snapshotMsg.UID)
	if !ok {
		return false
	}

	delete(list.idx, msgID)

	list.msg = append(
		list.msg[:index],
		list.msg[index+1:]...,
	)

	return true
}

func (list *snapMsgList) update(internalID imap.InternalMessageID, remoteID imap.MessageID) bool {
	snapMsg, ok := list.idx[internalID]
	if !ok {
		return false
	}

	snapMsg.ID.RemoteID = remoteID

	return true
}

func (list *snapMsgList) all() []*snapMsg {
	return list.msg
}

func (list *snapMsgList) len() int {
	return len(list.msg)
}

func (list *snapMsgList) where(fn func(seq snapMsgWithSeq) bool) []snapMsgWithSeq {
	var result []snapMsgWithSeq

	for idx, i := range list.msg {
		snapWithSeq := snapMsgWithSeq{
			snapMsg: i,
			Seq:     imap.SeqID(idx + 1),
		}

		if fn(snapWithSeq) {
			result = append(result, snapWithSeq)
		}
	}

	return result
}

func (list *snapMsgList) whereCount(fn func(seq snapMsgWithSeq) bool) int {
	result := 0

	for idx, i := range list.msg {
		snapWithSeq := snapMsgWithSeq{
			snapMsg: i,
			Seq:     imap.SeqID(idx + 1),
		}

		if fn(snapWithSeq) {
			result++
		}
	}

	return result
}

func (list *snapMsgList) has(msgID imap.InternalMessageID) bool {
	_, ok := list.idx[msgID]

	return ok
}

func (list *snapMsgList) get(msgID imap.InternalMessageID) (snapMsgWithSeq, bool) {
	snapshotMsg, ok := list.idx[msgID]
	if !ok {
		return snapMsgWithSeq{}, false
	}

	index, ok := list.binarySearchByUID(snapshotMsg.UID)
	if !ok {
		return snapMsgWithSeq{}, false
	}

	return snapMsgWithSeq{
		Seq:     imap.SeqID(index + 1),
		snapMsg: list.msg[index],
	}, ok
}

func (list *snapMsgList) seq(seq imap.SeqID) (snapMsgWithSeq, bool) {
	if imap.SeqID(len(list.msg)) < seq {
		return snapMsgWithSeq{}, false
	}

	return snapMsgWithSeq{
		Seq:     seq,
		snapMsg: list.msg[seq-1],
	}, true
}

func (list *snapMsgList) last() snapMsgWithSeq {
	return snapMsgWithSeq{
		Seq:     imap.SeqID(len(list.msg)),
		snapMsg: list.msg[len(list.msg)-1],
	}
}

func (list *snapMsgList) seqRange(seqLo, seqHi imap.SeqID) []snapMsgWithSeq {
	interval := list.msg[seqLo-1 : seqHi]
	result := make([]snapMsgWithSeq, len(interval))

	for i, v := range interval {
		result[i].Seq = imap.SeqID(int(seqLo) + i)
		result[i].snapMsg = v
	}

	return result
}

func (list *snapMsgList) uidRange(uidLo, uidHi imap.UID) []snapMsgWithSeq {
	listLen := len(list.msg)

	indexLo, _ := list.binarySearchByUID(uidLo)

	if indexLo >= listLen {
		return nil
	}

	indexHi, ok := list.binarySearchByUID(uidHi)
	if ok {
		indexHi++
	}

	if indexHi >= listLen {
		indexHi = listLen
	}

	interval := list.msg[indexLo:indexHi]
	result := make([]snapMsgWithSeq, len(interval))

	for i, v := range interval {
		result[i].Seq = imap.SeqID(indexLo + i + 1)
		result[i].snapMsg = v
	}

	return result
}

func (list *snapMsgList) getWithUID(uid imap.UID) (snapMsgWithSeq, bool) {
	index, ok := list.binarySearchByUID(uid)
	if !ok {
		return snapMsgWithSeq{}, false
	}

	return snapMsgWithSeq{
		Seq:     imap.SeqID(index + 1),
		snapMsg: list.msg[index],
	}, ok
}
