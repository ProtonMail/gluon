package state

import (
	"fmt"
	"slices"
	"sync"

	"github.com/ProtonMail/gluon/db"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/imap/command"
)

var ErrOutOfOrderUIDInsertion = fmt.Errorf("UIDs must be strictly ascending")

// snapMsg is a single message inside a snapshot.
type snapMsg struct {
	ID        db.MessageIDPair
	UID       imap.UID
	flags     imap.FlagSet
	toExpunge bool
}

type snapMsgWithSeq struct {
	Seq imap.SeqID
	*snapMsg
}

// snapMsgList is an ordered list of messages inside a snapshot.
type snapMsgList struct {
	msg  []*snapMsg
	idx  map[imap.InternalMessageID]*snapMsg
	lock sync.RWMutex
}

func newMsgList(capacity int) *snapMsgList {
	return &snapMsgList{
		msg: make([]*snapMsg, 0, capacity),
		idx: make(map[imap.InternalMessageID]*snapMsg, capacity),
	}
}

// withLock runs fn while holding the write lock. Every method that mutates msg/idx must go through this.
func (list *snapMsgList) withLock(fn func()) {
	list.lock.Lock()
	defer list.lock.Unlock()

	fn()
}

// withRLock runs fn while holding the read lock. Every method that reads msg/idx must go through this.
func (list *snapMsgList) withRLock(fn func()) {
	list.lock.RLock()
	defer list.lock.RUnlock()

	fn()
}

// binarySearchByUID assumes the caller already holds list.lock.
func (list *snapMsgList) binarySearchByUID(uid imap.UID) (int, bool) {
	msg := snapMsg{UID: uid}

	index, ok := slices.BinarySearchFunc(list.msg, &msg, func(s1 *snapMsg, s2 *snapMsg) int {
		return int(s1.UID) - int(s2.UID)
	})

	return index, ok
}

func (list *snapMsgList) insert(msgID db.MessageIDPair, msgUID imap.UID, flags imap.FlagSet) (err error) {
	list.withLock(func() {
		err = list.insertUnsafe(msgID, msgUID, flags)
	})

	return
}

func (list *snapMsgList) insertUnsafe(msgID db.MessageIDPair, msgUID imap.UID, flags imap.FlagSet) error {
	if len(list.msg) > 0 && list.msg[len(list.msg)-1].UID >= msgUID {
		return fmt.Errorf("UID-Last=%v UID-Msg=%v: %w", list.msg[len(list.msg)-1].UID, msgUID, ErrOutOfOrderUIDInsertion)
	}

	snapMsg := &snapMsg{
		ID:        msgID,
		UID:       msgUID,
		flags:     flags,
		toExpunge: flags.ContainsUnchecked(imap.FlagDeletedLowerCase),
	}

	list.msg = append(list.msg, snapMsg)
	list.idx[msgID.InternalID] = snapMsg

	return nil
}

func (list *snapMsgList) insertOutOfOrder(msgID db.MessageIDPair, msgUID imap.UID, flags imap.FlagSet) {
	list.withLock(func() {
		list.insertOutOfOrderUnsafe(msgID, msgUID, flags)
	})
}

func (list *snapMsgList) insertOutOfOrderUnsafe(msgID db.MessageIDPair, msgUID imap.UID, flags imap.FlagSet) {
	index, ok := list.binarySearchByUID(msgUID)
	if ok {
		panic("Duplicate UID added")
	}

	snapMsg := &snapMsg{
		ID:        msgID,
		UID:       msgUID,
		flags:     flags,
		toExpunge: flags.ContainsUnchecked(imap.FlagDeletedLowerCase),
	}

	list.msg = slices.Insert(list.msg, index, snapMsg)
	list.idx[msgID.InternalID] = snapMsg
}

func (list *snapMsgList) remove(msgID imap.InternalMessageID) (removed bool) {
	list.withLock(func() {
		removed = list.removeUnsafe(msgID)
	})

	return
}

func (list *snapMsgList) removeUnsafe(msgID imap.InternalMessageID) bool {
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

func (list *snapMsgList) update(internalID imap.InternalMessageID, remoteID imap.MessageID) (ok bool) {
	list.withRLock(func() {
		ok = list.updateUnsafe(internalID, remoteID)
	})

	return
}

func (list *snapMsgList) updateUnsafe(internalID imap.InternalMessageID, remoteID imap.MessageID) bool {
	snapMsg, ok := list.idx[internalID]
	if !ok {
		return false
	}

	snapMsg.ID.RemoteID = remoteID

	return true
}

func (list *snapMsgList) all() (result []*snapMsg) {
	list.withRLock(func() {
		result = list.allUnsafe()
	})

	return
}

func (list *snapMsgList) allUnsafe() []*snapMsg {
	return list.msg
}

func (list *snapMsgList) len() (result int) {
	list.withRLock(func() {
		result = list.lenUnsafe()
	})

	return
}

func (list *snapMsgList) lenUnsafe() int {
	return len(list.msg)
}

func (list *snapMsgList) where(fn func(seq snapMsgWithSeq) bool) (result []snapMsgWithSeq) {
	list.withRLock(func() {
		result = list.whereUnsafe(fn)
	})

	return
}

func (list *snapMsgList) whereUnsafe(fn func(seq snapMsgWithSeq) bool) []snapMsgWithSeq {
	var result []snapMsgWithSeq

	for idx, i := range list.msg {
		snapWithSeq := snapMsgWithSeq{
			snapMsg: i,
			Seq:     imap.SeqID(uint32(idx + 1)),
		}

		if fn(snapWithSeq) {
			result = append(result, snapWithSeq)
		}
	}

	return result
}

func (list *snapMsgList) whereCount(fn func(seq snapMsgWithSeq) bool) (result int) {
	list.withRLock(func() {
		result = list.whereCountUnsafe(fn)
	})

	return
}

func (list *snapMsgList) whereCountUnsafe(fn func(seq snapMsgWithSeq) bool) int {
	result := 0

	for idx, i := range list.msg {
		snapWithSeq := snapMsgWithSeq{
			snapMsg: i,
			Seq:     imap.SeqID(uint32(idx + 1)),
		}

		if fn(snapWithSeq) {
			result++
		}
	}

	return result
}

// firstWhere returns the first message (in sequence order) for which fn returns true.
func (list *snapMsgList) firstWhere(fn func(seq snapMsgWithSeq) bool) (result snapMsgWithSeq, ok bool) {
	list.withRLock(func() {
		result, ok = list.firstWhereUnsafe(fn)
	})

	return
}

func (list *snapMsgList) firstWhereUnsafe(fn func(seq snapMsgWithSeq) bool) (snapMsgWithSeq, bool) {
	for idx, i := range list.msg {
		snapWithSeq := snapMsgWithSeq{
			snapMsg: i,
			Seq:     imap.SeqID(uint32(idx + 1)),
		}

		if fn(snapWithSeq) {
			return snapWithSeq, true
		}
	}

	return snapMsgWithSeq{}, false
}

func (list *snapMsgList) has(msgID imap.InternalMessageID) (ok bool) {
	list.withRLock(func() {
		ok = list.hasUnsafe(msgID)
	})

	return
}

func (list *snapMsgList) hasUnsafe(msgID imap.InternalMessageID) bool {
	_, ok := list.idx[msgID]

	return ok
}

func (list *snapMsgList) get(msgID imap.InternalMessageID) (result snapMsgWithSeq, ok bool) {
	list.withRLock(func() {
		result, ok = list.getUnsafe(msgID)
	})

	return
}

func (list *snapMsgList) getUnsafe(msgID imap.InternalMessageID) (snapMsgWithSeq, bool) {
	snapshotMsg, ok := list.idx[msgID]
	if !ok {
		return snapMsgWithSeq{}, false
	}

	index, ok := list.binarySearchByUID(snapshotMsg.UID)
	if !ok {
		return snapMsgWithSeq{}, false
	}

	return snapMsgWithSeq{
		Seq:     imap.SeqID(uint32(index + 1)),
		snapMsg: list.msg[index],
	}, ok
}

func (list *snapMsgList) seq(seq imap.SeqID) (result snapMsgWithSeq, ok bool) {
	list.withRLock(func() {
		result, ok = list.seqUnsafe(seq)
	})

	return
}

func (list *snapMsgList) seqUnsafe(seq imap.SeqID) (snapMsgWithSeq, bool) {
	if imap.SeqID(uint32(len(list.msg))) < seq {
		return snapMsgWithSeq{}, false
	}

	return snapMsgWithSeq{
		Seq:     seq,
		snapMsg: list.msg[seq-1],
	}, true
}

func (list *snapMsgList) last() (result snapMsgWithSeq) {
	list.withRLock(func() {
		result = list.lastUnsafe()
	})

	return
}

func (list *snapMsgList) lastUnsafe() snapMsgWithSeq {
	return snapMsgWithSeq{
		Seq:     imap.SeqID(uint32(len(list.msg))),
		snapMsg: list.msg[len(list.msg)-1],
	}
}

func (list *snapMsgList) seqRange(seqLo, seqHi imap.SeqID) (result []snapMsgWithSeq) {
	list.withRLock(func() {
		result = list.seqRangeUnsafe(seqLo, seqHi)
	})

	return
}

func (list *snapMsgList) seqRangeUnsafe(seqLo, seqHi imap.SeqID) []snapMsgWithSeq {
	interval := list.msg[seqLo-1 : seqHi]
	result := make([]snapMsgWithSeq, len(interval))

	for i, v := range interval {
		result[i].Seq = imap.SeqID(uint32(seqLo) + uint32(i))
		result[i].snapMsg = v
	}

	return result
}

func (list *snapMsgList) uidRange(uidLo, uidHi imap.UID) (result []snapMsgWithSeq) {
	list.withRLock(func() {
		result = list.uidRangeUnsafe(uidLo, uidHi)
	})

	return
}

func (list *snapMsgList) uidRangeUnsafe(uidLo, uidHi imap.UID) []snapMsgWithSeq {
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
		result[i].Seq = imap.SeqID(uint32(indexLo + i + 1))
		result[i].snapMsg = v
	}

	return result
}

func (list *snapMsgList) getWithUID(uid imap.UID) (result snapMsgWithSeq, ok bool) {
	list.withRLock(func() {
		result, ok = list.getWithUIDUnsafe(uid)
	})

	return
}

func (list *snapMsgList) getWithUIDUnsafe(uid imap.UID) (snapMsgWithSeq, bool) {
	index, ok := list.binarySearchByUID(uid)
	if !ok {
		return snapMsgWithSeq{}, false
	}

	return snapMsgWithSeq{
		Seq:     imap.SeqID(uint32(index + 1)),
		snapMsg: list.msg[index],
	}, ok
}

func (list *snapMsgList) getWithSeqID(id imap.SeqID) (result snapMsgWithSeq, ok bool) {
	list.withRLock(func() {
		result, ok = list.getWithSeqIDUnsafe(id)
	})

	return
}

func (list *snapMsgList) getWithSeqIDUnsafe(id imap.SeqID) (snapMsgWithSeq, bool) {
	index := int(id) - 1
	listLen := len(list.msg)

	if listLen == 0 || index >= listLen {
		return snapMsgWithSeq{}, false
	}

	return snapMsgWithSeq{
		Seq:     id,
		snapMsg: list.msg[index],
	}, true
}

func (list *snapMsgList) existsWithSeqID(id imap.SeqID) (result bool) {
	list.withRLock(func() {
		result = list.existsWithSeqIDUnsafe(id)
	})

	return
}

func (list *snapMsgList) existsWithSeqIDUnsafe(id imap.SeqID) bool {
	index := int(id) - 1
	listLen := len(list.msg)

	return index < listLen
}

func (list *snapMsgList) resolveSeqInterval(seqSet []command.SeqRange) ([]SeqInterval, error) {
	res := make([]SeqInterval, 0, len(seqSet))

	for _, seqRange := range seqSet {
		if seqRange.Begin == seqRange.End {
			seq, err := list.resolveSeq(seqRange.Begin)
			if err != nil {
				return nil, err
			}

			res = append(res, SeqInterval{
				begin: seq,
				end:   seq,
			})
		} else {
			if seqRange.Begin == command.SeqNumValueAsterisk {
				seqRange.Begin, seqRange.End = seqRange.End, seqRange.Begin
			}

			begin, err := list.resolveSeq(seqRange.Begin)
			if err != nil {
				return nil, err
			}

			end, err := list.resolveSeq(seqRange.End)
			if err != nil {
				return nil, err
			}

			if begin > end {
				if seqRange.End != command.SeqNumValueAsterisk {
					begin, end = end, begin
				} else {
					end = begin
				}
			}

			res = append(res, SeqInterval{
				begin: begin,
				end:   end,
			})
		}
	}

	return res, nil
}

func (list *snapMsgList) resolveUIDInterval(seqSet []command.SeqRange) ([]UIDInterval, error) {
	res := make([]UIDInterval, 0, len(seqSet))

	for _, uidRange := range seqSet {
		if uidRange.Begin == uidRange.End {
			uid, err := list.resolveUID(uidRange.Begin)
			if err != nil {
				return nil, err
			}

			res = append(res, UIDInterval{
				begin: uid,
				end:   uid,
			})
		} else {
			if uidRange.Begin == command.SeqNumValueAsterisk {
				uidRange.Begin, uidRange.End = uidRange.End, uidRange.Begin
			}

			begin, err := list.resolveUID(uidRange.Begin)
			if err != nil {
				return nil, err
			}

			end, err := list.resolveUID(uidRange.End)
			if err != nil {
				return nil, err
			}

			if begin > end {
				if uidRange.End != command.SeqNumValueAsterisk {
					begin, end = end, begin
				} else {
					end = begin
				}
			}

			res = append(res, UIDInterval{
				begin: begin,
				end:   end,
			})

		}
	}

	return res, nil
}

// resolveSeq converts a textual sequence number into an integer.
// According to RFC 3501, the definition of seq-number, page 89, for message sequence numbers
// - No sequence number is valid if mailbox is empty, not even "*"
// - "*" is converted to the number of messages in the mailbox
// - when used in a range, the order of the indexes in irrelevant.
func (list *snapMsgList) resolveSeq(number command.SeqNum) (imap.SeqID, error) {
	if number == command.SeqNumValueAsterisk {
		return imap.SeqID(uint32(list.len())), nil
	}

	return imap.SeqID(uint32(number)), nil
}

// resolveUID converts a textual message UID into an integer.
func (list *snapMsgList) resolveUID(number command.SeqNum) (imap.UID, error) {
	if list.len() == 0 {
		return 0, ErrNoSuchMessage
	}

	if number == command.SeqNumValueAsterisk {
		return list.last().UID, nil
	}

	return imap.UID(uint32(number)), nil
}

func (list *snapMsgList) getMessagesInSeqRange(seqSet []command.SeqRange) ([]snapMsgWithSeq, error) {
	var res []snapMsgWithSeq

	intervals, err := list.resolveSeqInterval(seqSet)
	if err != nil {
		return nil, err
	}

	for _, seqRange := range intervals {
		if seqRange.begin == seqRange.end {
			msg, ok := list.getWithSeqID(seqRange.begin)
			if !ok {
				return nil, ErrNoSuchMessage
			}

			res = append(res, msg)
		} else {
			if !list.existsWithSeqID(seqRange.begin) || !list.existsWithSeqID(seqRange.end) {
				return nil, ErrNoSuchMessage
			}

			res = append(res, list.seqRange(seqRange.begin, seqRange.end)...)
		}
	}

	return res, nil
}

func (list *snapMsgList) getMessagesInUIDRange(seqSet []command.SeqRange) ([]snapMsgWithSeq, error) {
	var res []snapMsgWithSeq

	// If there are no messages in the mailbox, we still resolve without error.
	if list.len() == 0 {
		return nil, nil
	}

	intervals, err := list.resolveUIDInterval(seqSet)
	if err != nil {
		return nil, err
	}

	for _, uidRange := range intervals {
		if uidRange.begin == uidRange.end {
			msg, ok := list.getWithUID(uidRange.begin)
			if !ok {
				continue
			}

			res = append(res, msg)
		} else {
			res = append(res, list.uidRange(uidRange.begin, uidRange.end)...)
		}
	}

	return res, nil
}
