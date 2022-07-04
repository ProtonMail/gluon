package backend

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/backend/ent"
	"github.com/ProtonMail/gluon/internal/backend/ent/mailbox"
	"github.com/ProtonMail/gluon/internal/backend/ent/message"
	"github.com/ProtonMail/gluon/internal/backend/ent/uid"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/bradenaw/juniper/xslices"
)

type snapshot struct {
	mboxID string

	state    *State
	messages *snapMsgList
}

func newSnapshot(ctx context.Context, state *State, mbox *ent.Mailbox) (*snapshot, error) {
	snap := &snapshot{
		mboxID:   mbox.MailboxID,
		state:    state,
		messages: newMsgList(),
	}

	var msgUIDs []*ent.UID

	const limit = 16000

	for offset := 0; ; offset += limit {
		list, err := mbox.QueryUIDs().
			WithMessage(func(query *ent.MessageQuery) { query.WithFlags().Select(message.FieldMessageID) }).Offset(offset).Limit(limit).
			Select(uid.FieldUID, uid.FieldRecent, uid.FieldDeleted).
			All(ctx)

		if err != nil {
			return nil, err
		}

		if len(list) == 0 {
			break
		}

		msgUIDs = append(msgUIDs, list...)
	}

	for _, msgUID := range msgUIDs {
		snap.messages.insert(
			msgUID.Edges.Message.MessageID,
			msgUID.UID,
			newFlagSet(msgUID, msgUID.Edges.Message.Edges.Flags),
		)
	}

	return snap, nil
}

func (snap *snapshot) hasMessage(messageID string) bool {
	return snap.messages.has(messageID)
}

func (snap *snapshot) getMessage(messageID string) (*snapMsg, error) {
	msg, ok := snap.messages.get(messageID)
	if !ok {
		return nil, ErrNoSuchMessage
	}

	return msg, nil
}

func (snap *snapshot) getMessageBySeq(seq int) (*snapMsg, error) {
	msg, ok := snap.messages.seq(seq)
	if !ok {
		return nil, ErrNoSuchMessage
	}

	return msg, nil
}

func (snap *snapshot) getMessageSeq(messageID string) (int, error) {
	msg, ok := snap.messages.get(messageID)
	if !ok {
		return 0, ErrNoSuchMessage
	}

	return msg.Seq, nil
}

func (snap *snapshot) getMessageUID(messageID string) (int, error) {
	msg, ok := snap.messages.get(messageID)
	if !ok {
		return 0, ErrNoSuchMessage
	}

	return msg.UID, nil
}

func (snap *snapshot) getMessageFlags(messageID string) (imap.FlagSet, error) {
	msg, ok := snap.messages.get(messageID)
	if !ok {
		return nil, ErrNoSuchMessage
	}

	return msg.flags, nil
}

func (snap *snapshot) setMessageFlags(messageID string, flags imap.FlagSet) error {
	msg, ok := snap.messages.get(messageID)
	if !ok {
		return ErrNoSuchMessage
	}

	if recent := msg.flags.Contains(imap.FlagRecent); recent {
		flags = flags.Add(imap.FlagRecent)
	}

	msg.flags = flags

	return nil
}

func (snap *snapshot) getAllMessages() []*snapMsg {
	return snap.messages.all()
}

func (snap *snapshot) getAllMessageIDs() []string {
	return xslices.Map(snap.messages.all(), func(msg *snapMsg) string {
		return msg.ID
	})
}

func (snap *snapshot) getMessagesInRange(ctx context.Context, seq *proto.SequenceSet) ([]*snapMsg, error) {
	switch {
	case isUID(ctx):
		return snap.getMessagesInUIDRange(seq)

	default:
		return snap.getMessagesInSeqRange(seq)
	}
}

func (snap *snapshot) getMessagesInSeqRange(seq *proto.SequenceSet) ([]*snapMsg, error) {
	var res []*snapMsg

	for _, seqRange := range toSeqSet(seq) {
		switch len(seqRange) {
		case 1:
			seq, err := snap.resolveSeq(seqRange[0])
			if err != nil {
				return nil, err
			}

			res = append(res, snap.seqRange(seq, seq)...)

		case 2:
			begin, err := snap.resolveSeq(seqRange[0])
			if err != nil {
				return nil, err
			}

			end, err := snap.resolveSeq(seqRange[1])
			if err != nil {
				return nil, err
			}

			if begin > end {
				begin, end = end, begin
			}

			res = append(res, snap.seqRange(begin, end)...)

		default:
			panic("bad sequence range")
		}
	}

	return res, nil
}

func (snap *snapshot) getMessagesInUIDRange(seq *proto.SequenceSet) ([]*snapMsg, error) {
	var res []*snapMsg

	// If there are no messages in the mailbox, we still resolve without error.
	if snap.messages.len() == 0 {
		return nil, nil
	}

	for _, uidRange := range toSeqSet(seq) {
		switch len(uidRange) {
		case 1:
			uid, err := snap.resolveUID(uidRange[0])
			if err != nil {
				return nil, err
			}

			res = append(res, snap.uidRange(uid, uid)...)

		case 2:
			begin, err := snap.resolveUID(uidRange[0])
			if err != nil {
				return nil, err
			}

			end, err := snap.resolveUID(uidRange[1])
			if err != nil {
				return nil, err
			}

			if begin > end {
				begin, end = end, begin
			}

			res = append(res, snap.uidRange(begin, end)...)

		default:
			panic("bad sequence range")
		}
	}

	return res, nil
}

func (snap *snapshot) getMessagesWithFlag(flag string) []*snapMsg {
	return snap.messages.where(func(msg *snapMsg) bool {
		return msg.flags.Contains(flag)
	})
}

func (snap *snapshot) getMessagesWithoutFlag(flag string) []*snapMsg {
	return snap.messages.where(func(msg *snapMsg) bool {
		return !msg.flags.Contains(flag)
	})
}

func (snap *snapshot) appendMessage(ctx context.Context, tx *ent.Tx, messageID string) error {
	msgUID, err := tx.UID.Query().
		Where(
			uid.HasMailboxWith(mailbox.MailboxID(snap.mboxID)),
			uid.HasMessageWith(message.MessageID(messageID)),
		).
		WithMessage(func(query *ent.MessageQuery) { query.WithFlags() }).
		Only(ctx)
	if err != nil {
		return err
	}

	snap.messages.insert(
		messageID,
		msgUID.UID,
		newFlagSet(msgUID, msgUID.Edges.Message.Edges.Flags),
	)

	return nil
}

func (snap *snapshot) expungeMessage(ctx context.Context, tx *ent.Tx, messageID string) error {
	if ok := snap.messages.remove(messageID); !ok {
		return ErrNoSuchMessage
	}

	return nil
}

func (snap *snapshot) updateMailboxID(oldID, newID string) error {
	if snap.mboxID != oldID {
		return ErrNoSuchMailbox
	}

	snap.mboxID = newID

	return nil
}

func (snap *snapshot) updateMessageID(oldID, newID string) error {
	if ok := snap.messages.update(oldID, newID); !ok {
		return ErrNoSuchMessage
	}

	return nil
}

// TODO: How serious is the performance impact of this?
func (snap *snapshot) seqRange(seqLo, seqHi int) []*snapMsg {
	return snap.messages.where(func(msg *snapMsg) bool {
		return seqLo <= msg.Seq && msg.Seq <= seqHi
	})
}

// TODO: How serious is the performance impact of this?
func (snap *snapshot) uidRange(uidLo, uidHi int) []*snapMsg {
	return snap.messages.where(func(msg *snapMsg) bool {
		return uidLo <= msg.UID && msg.UID <= uidHi
	})
}

// resolveSeq converts a textual sequence number into an integer.
// According to RFC 3501, the definition of seq-number, page 89, for message sequence numbers
// - No sequence number is valid if mailbox is empty, not even "*"
// - "*" is converted to the number of messages in the mailbox
// - when used in a range, the order of the indexes in irrelevant.
func (snap *snapshot) resolveSeq(number string) (int, error) {
	if snap.messages.len() == 0 {
		return 0, ErrNoSuchMessage
	}

	if number == "*" {
		return snap.messages.last().Seq, nil
	}

	seq, err := strconv.Atoi(number)
	if err != nil {
		return 0, fmt.Errorf("failed to parse sequence number: %w", err)
	}

	msg, ok := snap.messages.seq(seq)
	if !ok {
		return 0, ErrNoSuchMessage
	}

	return msg.Seq, nil
}

// resolveUID converts a textual message UID into an integer.
func (snap *snapshot) resolveUID(number string) (int, error) {
	if snap.messages.len() == 0 {
		return 0, ErrNoSuchMessage
	}

	if number == "*" {
		return snap.messages.last().UID, nil
	}

	return strconv.Atoi(number)
}

func newFlagSet(msgUID *ent.UID, flags []*ent.MessageFlag) imap.FlagSet {
	flagSet := imap.NewFlagSetFromSlice(xslices.Map(flags, func(flag *ent.MessageFlag) string {
		return flag.Value
	}))

	if msgUID.Deleted {
		flagSet = flagSet.Add(imap.FlagDeleted)
	}

	if msgUID.Recent {
		flagSet = flagSet.Add(imap.FlagRecent)
	}

	return flagSet
}
