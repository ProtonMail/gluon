package backend

import (
	"context"
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/backend/ent"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/rfc822"
	"github.com/bradenaw/juniper/xslices"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
)

type Mailbox struct {
	mbox *ent.Mailbox

	state *State
	snap  *snapshotWrapper

	selected bool
	readOnly bool
}

func newMailbox(mbox *ent.Mailbox, state *State, wrapper *snapshotWrapper) *Mailbox {
	selected := snapshotRead(wrapper, func(s *snapshot) bool {
		return s != nil
	})

	return &Mailbox{
		mbox: mbox,

		state: state,

		selected: selected,
		readOnly: state.ro,
		snap:     wrapper,
	}
}

func (m *Mailbox) Name() string {
	return m.mbox.Name
}

func (m *Mailbox) Selected() bool {
	return m.selected
}

func (m *Mailbox) ReadOnly() bool {
	return m.readOnly
}

func (m *Mailbox) ExpungeIssued() bool {
	var issued bool

	for _, res := range m.state.res {
		switch res.(type) {
		case *expunge:
			issued = true
		}
	}

	return issued
}

func (m *Mailbox) Count() int {
	return snapshotRead(m.snap, func(s *snapshot) int {
		return len(s.getAllMessages())
	})
}

func (m *Mailbox) Flags(ctx context.Context) (imap.FlagSet, error) {
	flags, err := m.mbox.QueryFlags().All(ctx)
	if err != nil {
		return nil, err
	}

	return imap.NewFlagSetFromSlice(xslices.Map(flags, func(flag *ent.MailboxFlag) string {
		return flag.Value
	})), nil
}

func (m *Mailbox) PermanentFlags(ctx context.Context) (imap.FlagSet, error) {
	permFlags, err := m.mbox.QueryPermanentFlags().All(ctx)
	if err != nil {
		return nil, err
	}

	return imap.NewFlagSetFromSlice(xslices.Map(permFlags, func(flag *ent.MailboxPermFlag) string {
		return flag.Value
	})), nil
}

func (m *Mailbox) Attributes(ctx context.Context) (imap.FlagSet, error) {
	attrs, err := m.mbox.QueryAttributes().All(ctx)
	if err != nil {
		return nil, err
	}

	return imap.NewFlagSetFromSlice(xslices.Map(attrs, func(flag *ent.MailboxAttr) string {
		return flag.Value
	})), nil
}

func (m *Mailbox) UIDNext() int {
	return m.mbox.UIDNext
}

func (m *Mailbox) UIDValidity() int {
	return m.mbox.UIDValidity
}

func (m *Mailbox) Subscribed() bool {
	return m.mbox.Subscribed
}

func (m *Mailbox) GetMessagesWithFlag(flag string) []int {
	return snapshotRead(m.snap, func(s *snapshot) []int {
		return xslices.Map(s.getMessagesWithFlag(flag), func(msg *snapMsg) int {
			return msg.Seq
		})
	})
}

func (m *Mailbox) GetMessagesWithoutFlag(flag string) []int {
	return snapshotRead(m.snap, func(s *snapshot) []int {
		return xslices.Map(s.getMessagesWithoutFlag(flag), func(msg *snapMsg) int {
			return msg.Seq
		})
	})
}

func (m *Mailbox) Append(ctx context.Context, literal []byte, flags imap.FlagSet, date time.Time) (int, error) {
	internalID, err := rfc822.GetHeaderValue(literal, InternalIDKey)
	if err != nil {
		return 0, err
	}

	if len(internalID) > 0 {
		msgID := imap.InternalMessageID(internalID)

		if exists, err := DBReadResult(ctx, m.state.db, func(ctx context.Context, client *ent.Client) (bool, error) {
			return DBHasMessageWithID(ctx, client, msgID)
		}); err != nil || !exists {
			logrus.WithError(err).Warn("The message has an unknown internal ID")
		} else if res, err := DBWriteResult(ctx, m.state.db, func(ctx context.Context, tx *ent.Tx) (map[imap.InternalMessageID]int, error) {
			return m.state.actionAddMessagesToMailbox(ctx, tx, []MessageIDPair{NewMessageIDPairWithoutRemote(msgID)}, NewMailboxIDPair(m.mbox))
		}); err != nil {
			return 0, err
		} else {
			return res[msgID], nil
		}
	}

	snapMBoxID := snapshotRead(m.snap, func(s *snapshot) MailboxIDPair {
		return s.mboxID
	})

	return DBWriteResult(ctx, m.state.db, func(ctx context.Context, tx *ent.Tx) (int, error) {
		return m.state.actionCreateMessage(ctx, tx, snapMBoxID, literal, flags, date)
	})
}

// Copy copies the messages represented by the given sequence set into the mailbox with the given name.
// If the context is a UID context, the sequence set refers to message UIDs.
// If no items are copied the response object will be nil.
func (m *Mailbox) Copy(ctx context.Context, seq *proto.SequenceSet, name string) (response.Item, error) {
	mbox, err := DBReadResult(ctx, m.state.db, func(ctx context.Context, client *ent.Client) (*ent.Mailbox, error) {
		return DBGetMailboxByName(ctx, client, name)
	})
	if err != nil {
		return nil, ErrNoSuchMailbox
	}

	messages, err := snapshotReadErr(m.snap, func(s *snapshot) ([]*snapMsg, error) {
		return s.getMessagesInRange(ctx, seq)
	})
	if err != nil {
		return nil, err
	}

	msgIDs := xslices.Map(messages, func(msg *snapMsg) MessageIDPair {
		return msg.ID
	})

	msgUIDs := xslices.Map(messages, func(msg *snapMsg) int {
		return msg.UID
	})

	destUIDs, err := DBWriteResult(ctx, m.state.db, func(ctx context.Context, tx *ent.Tx) (map[imap.InternalMessageID]int, error) {
		return m.state.actionAddMessagesToMailbox(ctx, tx, msgIDs, NewMailboxIDPair(mbox))
	})
	if err != nil {
		return nil, err
	}

	var res response.Item

	if len(destUIDs) > 0 {
		res = response.ItemCopyUID(mbox.UIDValidity, msgUIDs, xslices.Map(maps.Keys(destUIDs), func(messageID imap.InternalMessageID) int {
			return destUIDs[messageID]
		}))
	}

	return res, nil
}

// Move moves the messages represented by the given sequence set into the mailbox with the given name.
// If the context is a UID context, the sequence set refers to message UIDs.
// If no items are moved the response object will be nil.
func (m *Mailbox) Move(ctx context.Context, seq *proto.SequenceSet, name string) (response.Item, error) {
	mbox, err := DBReadResult(ctx, m.state.db, func(ctx context.Context, client *ent.Client) (*ent.Mailbox, error) {
		return DBGetMailboxByName(ctx, client, name)
	})

	if err != nil {
		return nil, ErrNoSuchMailbox
	}

	var snapMBoxID MailboxIDPair

	messages, err := snapshotReadErr(m.snap, func(s *snapshot) ([]*snapMsg, error) {
		snapMBoxID = s.mboxID
		return s.getMessagesInRange(ctx, seq)
	})
	if err != nil {
		return nil, err
	}

	msgIDs := xslices.Map(messages, func(msg *snapMsg) MessageIDPair {
		return msg.ID
	})

	msgUIDs := xslices.Map(messages, func(msg *snapMsg) int {
		return msg.UID
	})

	destUIDs, err := DBWriteResult(ctx, m.state.db, func(ctx context.Context, tx *ent.Tx) (map[imap.InternalMessageID]int, error) {
		return m.state.actionMoveMessages(ctx, tx, msgIDs, snapMBoxID, NewMailboxIDPair(mbox))
	})
	if err != nil {
		return nil, err
	}

	var res response.Item

	if len(destUIDs) > 0 {
		res = response.ItemCopyUID(mbox.UIDValidity, msgUIDs, xslices.Map(maps.Keys(destUIDs), func(messageID imap.InternalMessageID) int {
			return destUIDs[messageID]
		}))
	}

	return res, nil
}

func (m *Mailbox) Store(ctx context.Context, seq *proto.SequenceSet, operation proto.Operation, flags imap.FlagSet) error {
	messages, err := snapshotReadErr(m.snap, func(s *snapshot) ([]*snapMsg, error) {
		return s.getMessagesInRange(ctx, seq)
	})
	if err != nil {
		return err
	}

	msgIDs := xslices.Map(messages, func(msg *snapMsg) MessageIDPair {
		return msg.ID
	})

	return m.state.db.Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		switch operation {
		case proto.Operation_Add:
			if _, err := m.state.actionAddMessageFlags(ctx, tx, msgIDs, flags); err != nil {
				return err
			}

		case proto.Operation_Remove:
			if _, err := m.state.actionRemoveMessageFlags(ctx, tx, msgIDs, flags); err != nil {
				return err
			}

		case proto.Operation_Replace:
			if err := m.state.actionSetMessageFlags(ctx, tx, msgIDs, flags); err != nil {
				return err
			}
		}

		return nil
	})
}

func (m *Mailbox) Expunge(ctx context.Context, seq *proto.SequenceSet) error {
	var msg []*snapMsg

	if seq != nil {
		snapMsgs, err := snapshotReadErr(m.snap, func(s *snapshot) ([]*snapMsg, error) {
			return s.getMessagesInRange(ctx, seq)
		})
		if err != nil {
			return err
		}

		msg = snapMsgs
	} else {
		msg = snapshotRead(m.snap, func(s *snapshot) []*snapMsg {
			return s.getAllMessages()
		})
	}

	return m.expunge(ctx, msg)
}

func (m *Mailbox) expunge(ctx context.Context, messages []*snapMsg) error {
	messages = xslices.Filter(messages, func(msg *snapMsg) bool {
		return msg.flags.Contains(imap.FlagDeleted)
	})

	msgIDs := xslices.Map(messages, func(msg *snapMsg) MessageIDPair {
		return msg.ID
	})

	mboxID := snapshotRead(m.snap, func(s *snapshot) MailboxIDPair {
		return s.mboxID
	})

	return m.state.db.Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		return m.state.actionRemoveMessagesFromMailbox(ctx, tx, msgIDs, mboxID)
	})
}

func (m *Mailbox) Flush(ctx context.Context, permitExpunge bool) ([]response.Response, error) {
	return DBWriteResult(ctx, m.state.db, func(ctx context.Context, tx *ent.Tx) ([]response.Response, error) {
		return m.state.flushResponses(ctx, tx, permitExpunge)
	})
}

func (m *Mailbox) Close(ctx context.Context) error {
	if err := m.state.deleteConnMetadata(); err != nil {
		return err
	}

	return m.state.close()
}
