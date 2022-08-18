package backend

import (
	"context"
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/backend/ent"
	"github.com/ProtonMail/gluon/internal/backend/ent/mailbox"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/rfc822"
	"github.com/bradenaw/juniper/xslices"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
)

type Mailbox struct {
	tx   *ent.Tx
	mbox *ent.Mailbox

	state *State
	snap  *snapshot

	selected bool
	readOnly bool
}

func newMailbox(tx *ent.Tx, mbox *ent.Mailbox, state *State, snap *snapshot) *Mailbox {
	return &Mailbox{
		tx:   tx,
		mbox: mbox,

		state: state,
		snap:  snap,

		selected: state.snap != nil,
		readOnly: state.ro,
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
	return len(m.snap.getAllMessages())
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
	return xslices.Map(m.snap.getMessagesWithFlag(flag), func(msg *snapMsg) int {
		return msg.Seq
	})
}

func (m *Mailbox) GetMessagesWithoutFlag(flag string) []int {
	return xslices.Map(m.snap.getMessagesWithoutFlag(flag), func(msg *snapMsg) int {
		return msg.Seq
	})
}

func (m *Mailbox) Append(ctx context.Context, literal []byte, flags imap.FlagSet, date time.Time) (int, error) {
	internalID, err := rfc822.GetHeaderValue(literal, InternalIDKey)
	if err != nil {
		return 0, err
	}

	if len(internalID) > 0 {
		msgID := imap.InternalMessageID(internalID)
		if exists, err := DBHasMessageWithID(ctx, m.tx.Client(), msgID); err != nil || !exists {
			logrus.WithError(err).Warn("The message has an unknown internal ID")
		} else if res, err := m.state.actionAddMessagesToMailbox(ctx, m.tx, []MessageIDPair{NewMessageIDPairWithoutRemote(msgID)}, NewMailboxIDPair(m.mbox)); err != nil {
			return 0, err
		} else {
			return res[msgID], nil
		}
	}

	return m.state.actionCreateMessage(ctx, m.tx, m.snap.mboxID, literal, flags, date)
}

// Copy copies the messages represented by the given sequence set into the mailbox with the given name.
// If the context is a UID context, the sequence set refers to message UIDs.
// If no items are copied the response object will be nil.
func (m *Mailbox) Copy(ctx context.Context, seq *proto.SequenceSet, name string) (response.Item, error) {
	mbox, err := m.tx.Mailbox.Query().Where(mailbox.Name(name)).Only(ctx)
	if err != nil {
		return nil, ErrNoSuchMailbox
	}

	messages, err := m.snap.getMessagesInRange(ctx, seq)
	if err != nil {
		return nil, err
	}

	msgIDs := xslices.Map(messages, func(msg *snapMsg) MessageIDPair {
		return msg.ID
	})

	msgUIDs := xslices.Map(messages, func(msg *snapMsg) int {
		return msg.UID
	})

	destUIDs, err := m.state.actionAddMessagesToMailbox(ctx, m.tx, msgIDs, NewMailboxIDPair(mbox))
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
	mbox, err := m.tx.Mailbox.Query().Where(mailbox.Name(name)).Only(ctx)
	if err != nil {
		return nil, ErrNoSuchMailbox
	}

	messages, err := m.snap.getMessagesInRange(ctx, seq)
	if err != nil {
		return nil, err
	}

	msgIDs := xslices.Map(messages, func(msg *snapMsg) MessageIDPair {
		return msg.ID
	})

	msgUIDs := xslices.Map(messages, func(msg *snapMsg) int {
		return msg.UID
	})

	destUIDs, err := m.state.actionMoveMessages(ctx, m.tx, msgIDs, m.snap.mboxID, NewMailboxIDPair(mbox))
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
	messages, err := m.snap.getMessagesInRange(ctx, seq)
	if err != nil {
		return err
	}

	msgIDs := xslices.Map(messages, func(msg *snapMsg) MessageIDPair {
		return msg.ID
	})

	switch operation {
	case proto.Operation_Add:
		if _, err := m.state.actionAddMessageFlags(ctx, m.tx, msgIDs, flags); err != nil {
			return err
		}

	case proto.Operation_Remove:
		if _, err := m.state.actionRemoveMessageFlags(ctx, m.tx, msgIDs, flags); err != nil {
			return err
		}

	case proto.Operation_Replace:
		if err := m.state.actionSetMessageFlags(ctx, m.tx, msgIDs, flags); err != nil {
			return err
		}
	}

	return nil
}

func (m *Mailbox) Expunge(ctx context.Context, seq *proto.SequenceSet) error {
	var msg []*snapMsg

	if seq != nil {
		var err error

		if msg, err = m.snap.getMessagesInRange(ctx, seq); err != nil {
			return err
		}
	} else {
		msg = m.snap.getAllMessages()
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

	return m.state.actionRemoveMessagesFromMailbox(ctx, m.tx, msgIDs, m.snap.mboxID)
}

func (m *Mailbox) Flush(ctx context.Context, permitExpunge bool) ([]response.Response, error) {
	return m.state.flushResponses(ctx, m.tx, permitExpunge)
}

func (m *Mailbox) Close(ctx context.Context) error {
	if err := m.state.deleteConnMetadata(); err != nil {
		return err
	}

	return m.state.close(ctx, m.tx)
}
