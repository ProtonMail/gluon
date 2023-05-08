package state

import (
	"context"
	"strings"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/imap/command"
	"github.com/ProtonMail/gluon/internal/contexts"
	"github.com/ProtonMail/gluon/internal/db"
	"github.com/ProtonMail/gluon/internal/db/ent"
	"github.com/ProtonMail/gluon/internal/ids"
	"github.com/bradenaw/juniper/xslices"
)

type snapshot struct {
	mboxID ids.MailboxIDPair

	state    *State
	messages *snapMsgList
}

func newSnapshot(ctx context.Context, state *State, client *ent.Client, mbox *ent.Mailbox) (*snapshot, error) {
	snapshotMessages, err := db.GetMailboxMessagesForNewSnapshot(ctx, client, mbox.ID)
	if err != nil {
		return nil, err
	}

	snap := &snapshot{
		mboxID:   ids.NewMailboxIDPair(mbox),
		state:    state,
		messages: newMsgList(len(snapshotMessages)),
	}

	for _, snapshotMessage := range snapshotMessages {
		if err := snap.messages.insert(
			ids.MessageIDPair{InternalID: snapshotMessage.InternalID, RemoteID: snapshotMessage.RemoteID},
			snapshotMessage.UID,
			snapshotMessage.GetFlagSet(),
		); err != nil {
			return nil, err
		}
	}

	return snap, nil
}

func newEmptySnapshot(state *State, mbox *ent.Mailbox) *snapshot {
	return &snapshot{
		mboxID:   ids.NewMailboxIDPair(mbox),
		state:    state,
		messages: newMsgList(0),
	}
}

func (snap *snapshot) len() int {
	return snap.messages.len()
}

func (snap *snapshot) hasMessage(messageID imap.InternalMessageID) bool {
	return snap.messages.has(messageID)
}

func (snap *snapshot) getMessageSeq(messageID imap.InternalMessageID) (imap.SeqID, error) {
	msg, ok := snap.messages.get(messageID)
	if !ok {
		return 0, ErrNoSuchMessage
	}

	return msg.Seq, nil
}

func (snap *snapshot) getMessageUID(messageID imap.InternalMessageID) (imap.UID, error) {
	msg, ok := snap.messages.get(messageID)
	if !ok {
		return 0, ErrNoSuchMessage
	}

	return msg.UID, nil
}

func (snap *snapshot) getMessageFlags(messageID imap.InternalMessageID) (imap.FlagSet, error) {
	msg, ok := snap.messages.get(messageID)
	if !ok {
		return nil, ErrNoSuchMessage
	}

	return msg.flags, nil
}

func (snap *snapshot) setMessageFlags(messageID imap.InternalMessageID, flags imap.FlagSet) error {
	msg, ok := snap.messages.get(messageID)
	if !ok {
		return ErrNoSuchMessage
	}

	if recent := msg.flags.ContainsUnchecked(imap.FlagRecentLowerCase); recent {
		flags = flags.Add(imap.FlagRecent)
	}

	msg.toExpunge = flags.ContainsUnchecked(imap.FlagDeletedLowerCase)

	msg.flags = flags

	return nil
}

func (snap *snapshot) getAllMessages() []snapMsgWithSeq {
	allMessages := snap.messages.all()
	result := make([]snapMsgWithSeq, len(allMessages))

	for i, v := range allMessages {
		result[i] = snapMsgWithSeq{
			Seq:     imap.SeqID(i + 1),
			snapMsg: v,
		}
	}

	return result
}

func (snap *snapshot) getAllMessageIDs() []ids.MessageIDPair {
	return xslices.Map(snap.messages.all(), func(msg *snapMsg) ids.MessageIDPair {
		return msg.ID
	})
}

func (snap *snapshot) getAllMessagesIDsMarkedDelete() []ids.MessageIDPair {
	var msgs []ids.MessageIDPair

	for _, v := range snap.messages.all() {
		if v.toExpunge {
			msgs = append(msgs, v.ID)
		}
	}

	return msgs
}

func (snap *snapshot) getMessagesInRange(ctx context.Context, seq []command.SeqRange) ([]snapMsgWithSeq, error) {
	switch {
	case contexts.IsUID(ctx):
		return snap.getMessagesInUIDRange(seq)

	default:
		return snap.getMessagesInSeqRange(seq)
	}
}

type SeqInterval struct {
	begin imap.SeqID
	end   imap.SeqID
}

func (s SeqInterval) contains(seq imap.SeqID) bool {
	return seq >= s.begin && seq <= s.end
}

type UIDInterval struct {
	begin imap.UID
	end   imap.UID
}

func (u UIDInterval) contains(uid imap.UID) bool {
	return uid >= u.begin && uid <= u.end
}

func (snap *snapshot) resolveSeqInterval(seq []command.SeqRange) ([]SeqInterval, error) {
	return snap.messages.resolveSeqInterval(seq)
}

func (snap *snapshot) resolveUIDInterval(seq []command.SeqRange) ([]UIDInterval, error) {
	return snap.messages.resolveUIDInterval(seq)
}

func (snap *snapshot) getMessagesInSeqRange(seq []command.SeqRange) ([]snapMsgWithSeq, error) {
	return snap.messages.getMessagesInSeqRange(seq)
}

func (snap *snapshot) getMessagesInUIDRange(seq []command.SeqRange) ([]snapMsgWithSeq, error) {
	return snap.messages.getMessagesInUIDRange(seq)
}

func (snap *snapshot) firstMessageWithFlag(flag string) (snapMsgWithSeq, bool) {
	flagLower := strings.ToLower(flag)

	for i, msg := range snap.messages.msg {
		if msg.flags.ContainsUnchecked(flagLower) {
			return snapMsgWithSeq{Seq: imap.SeqID(i + 1), snapMsg: msg}, true
		}
	}

	return snapMsgWithSeq{}, false
}

func (snap *snapshot) firstMessageWithoutFlag(flag string) (snapMsgWithSeq, bool) {
	flagLower := strings.ToLower(flag)

	for i, msg := range snap.messages.msg {
		if !msg.flags.ContainsUnchecked(flagLower) {
			return snapMsgWithSeq{Seq: imap.SeqID(i + 1), snapMsg: msg}, true
		}
	}

	return snapMsgWithSeq{}, false
}

func (snap *snapshot) getMessagesWithFlag(flag string) []snapMsgWithSeq {
	flagLower := strings.ToLower(flag)

	return snap.messages.where(func(msg snapMsgWithSeq) bool {
		return msg.flags.ContainsUnchecked(flagLower)
	})
}

func (snap *snapshot) getMessagesWithFlagCount(flag string) int {
	flagLower := strings.ToLower(flag)

	return snap.messages.whereCount(func(msg snapMsgWithSeq) bool {
		return msg.flags.ContainsUnchecked(flagLower)
	})
}

func (snap *snapshot) getMessagesWithoutFlag(flag string) []snapMsgWithSeq {
	flagLower := strings.ToLower(flag)

	return snap.messages.where(func(msg snapMsgWithSeq) bool {
		return !msg.flags.ContainsUnchecked(flagLower)
	})
}

func (snap *snapshot) getMessagesWithoutFlagCount(flag string) int {
	flagLower := strings.ToLower(flag)

	return snap.messages.whereCount(func(msg snapMsgWithSeq) bool {
		return !msg.flags.ContainsUnchecked(flagLower)
	})
}

func (snap *snapshot) appendMessage(messageID ids.MessageIDPair, uid imap.UID, flags imap.FlagSet) error {
	return snap.messages.insert(
		messageID,
		uid,
		flags,
	)
}

func (snap *snapshot) appendMessageFromOtherState(messageID ids.MessageIDPair, uid imap.UID, flags imap.FlagSet) error {
	snap.messages.insertOutOfOrder(
		messageID,
		uid,
		flags,
	)

	return nil
}

func (snap *snapshot) expungeMessage(messageID imap.InternalMessageID) error {
	if ok := snap.messages.remove(messageID); !ok {
		return ErrNoSuchMessage
	}

	return nil
}

func (snap *snapshot) updateMailboxRemoteID(internalID imap.InternalMailboxID, remoteID imap.MailboxID) error {
	if snap.mboxID.InternalID != internalID {
		return ErrNoSuchMailbox
	}

	snap.mboxID.RemoteID = remoteID

	return nil
}

func (snap *snapshot) updateMessageRemoteID(internalID imap.InternalMessageID, remoteID imap.MessageID) error {
	if ok := snap.messages.update(internalID, remoteID); !ok {
		return ErrNoSuchMessage
	}

	return nil
}
