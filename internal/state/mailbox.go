package state

import (
	"context"
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db"
	"github.com/ProtonMail/gluon/internal/db/ent"
	"github.com/ProtonMail/gluon/internal/ids"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/rfc822"
	"github.com/ProtonMail/gluon/store"
	"github.com/bradenaw/juniper/xslices"
	"github.com/sirupsen/logrus"
)

type Mailbox struct {
	mbox *ent.Mailbox

	state *State
	snap  *snapshot

	selected bool
	readOnly bool
}

type AppendOnlyMailbox interface {
	Append(ctx context.Context, literal []byte, flags imap.FlagSet, date time.Time) (imap.UID, error)
	Flush(ctx context.Context, permitExpunge bool) ([]response.Response, error)
	UIDValidity() imap.UID
}

func newMailbox(mbox *ent.Mailbox, state *State, snap *snapshot) *Mailbox {
	return &Mailbox{
		mbox: mbox,

		state: state,

		selected: snap != nil,
		readOnly: state.ro,
		snap:     snap,
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

func (m *Mailbox) UIDNext() imap.UID {
	return m.mbox.UIDNext
}

func (m *Mailbox) UIDValidity() imap.UID {
	return m.mbox.UIDValidity
}

func (m *Mailbox) Subscribed() bool {
	return m.mbox.Subscribed
}

func (m *Mailbox) GetMessagesWithFlag(flag string) []imap.SeqID {
	return xslices.Map(m.snap.getMessagesWithFlag(flag), func(msg *snapMsg) imap.SeqID {
		return msg.Seq
	})
}

func (m *Mailbox) GetMessagesWithoutFlag(flag string) []imap.SeqID {
	return xslices.Map(m.snap.getMessagesWithoutFlag(flag), func(msg *snapMsg) imap.SeqID {
		return msg.Seq
	})
}

func (m *Mailbox) Append(ctx context.Context, literal []byte, flags imap.FlagSet, date time.Time) (imap.UID, error) {
	internalIDString, err := rfc822.GetHeaderValue(literal, ids.InternalIDKey)
	if err != nil {
		return 0, err
	}

	if len(internalIDString) > 0 {
		msgID, err := imap.InternalMessageIDFromString(internalIDString)
		if err != nil {
			return 0, err
		}

		if exists, err := db.ReadResult(ctx, m.state.db(), func(ctx context.Context, client *ent.Client) (bool, error) {
			return db.HasMessageWithID(ctx, client, msgID)
		}); err != nil || !exists {
			logrus.WithError(err).Warn("The message has an unknown internal ID")
		} else if res, err := db.WriteResult(ctx, m.state.db(), func(ctx context.Context, tx *ent.Tx) ([]db.UIDWithFlags, error) {
			return m.state.actionAddMessagesToMailbox(ctx, tx, []ids.MessageIDPair{ids.NewMessageIDPairWithoutRemote(msgID)}, ids.NewMailboxIDPair(m.mbox), m.snap == m.state.snap)
		}); err != nil {
			return 0, err
		} else {
			return res[0].UID, nil
		}
	}

	return db.WriteAndStoreResult(ctx, m.state.db(), m.state.user.GetStore(), func(ctx context.Context, tx *ent.Tx, transaction store.Transaction) (imap.UID, error) {
		return m.state.actionCreateMessage(ctx, tx, transaction, m.snap.mboxID, literal, flags, date, m.snap == m.state.snap)
	})
}

// Copy copies the messages represented by the given sequence set into the mailbox with the given name.
// If the context is a UID context, the sequence set refers to message UIDs.
// If no items are copied the response object will be nil.
func (m *Mailbox) Copy(ctx context.Context, seq *proto.SequenceSet, name string) (response.Item, error) {
	mbox, err := db.ReadResult(ctx, m.state.db(), func(ctx context.Context, client *ent.Client) (*ent.Mailbox, error) {
		return db.GetMailboxByName(ctx, client, name)
	})
	if err != nil {
		return nil, ErrNoSuchMailbox
	}

	messages, err := m.snap.getMessagesInRange(ctx, seq)
	if err != nil {
		return nil, err
	}

	msgIDs := make([]ids.MessageIDPair, len(messages))
	msgUIDs := make([]imap.UID, len(messages))

	for i := 0; i < len(messages); i++ {
		snapMsg := messages[i]
		msgUIDs[i] = snapMsg.UID
		msgIDs[i] = snapMsg.ID
	}

	destUIDs, err := db.WriteResult(ctx, m.state.db(), func(ctx context.Context, tx *ent.Tx) ([]db.UIDWithFlags, error) {
		return m.state.actionAddMessagesToMailbox(ctx, tx, msgIDs, ids.NewMailboxIDPair(mbox), m.snap == m.state.snap)
	})
	if err != nil {
		return nil, err
	}

	var res response.Item

	if len(destUIDs) > 0 {
		res = response.ItemCopyUID(mbox.UIDValidity, msgUIDs, xslices.Map(destUIDs, func(uid db.UIDWithFlags) imap.UID {
			return uid.UID
		}))
	}

	return res, nil
}

// Move moves the messages represented by the given sequence set into the mailbox with the given name.
// If the context is a UID context, the sequence set refers to message UIDs.
// If no items are moved the response object will be nil.
func (m *Mailbox) Move(ctx context.Context, seq *proto.SequenceSet, name string) (response.Item, error) {
	mbox, err := db.ReadResult(ctx, m.state.db(), func(ctx context.Context, client *ent.Client) (*ent.Mailbox, error) {
		return db.GetMailboxByName(ctx, client, name)
	})
	if err != nil {
		return nil, ErrNoSuchMailbox
	}

	messages, err := m.snap.getMessagesInRange(ctx, seq)
	if err != nil {
		return nil, err
	}

	msgIDs := make([]ids.MessageIDPair, len(messages))
	msgUIDs := make([]imap.UID, len(messages))

	for i := 0; i < len(messages); i++ {
		snapMsg := messages[i]
		msgUIDs[i] = snapMsg.UID
		msgIDs[i] = snapMsg.ID
	}

	destUIDs, err := db.WriteResult(ctx, m.state.db(), func(ctx context.Context, tx *ent.Tx) ([]db.UIDWithFlags, error) {
		return m.state.actionMoveMessages(ctx, tx, msgIDs, m.snap.mboxID, ids.NewMailboxIDPair(mbox))
	})
	if err != nil {
		return nil, err
	}

	var res response.Item

	if len(destUIDs) > 0 {
		res = response.ItemCopyUID(mbox.UIDValidity, msgUIDs, xslices.Map(destUIDs, func(uid db.UIDWithFlags) imap.UID {
			return uid.UID
		}))
	}

	return res, nil
}

func (m *Mailbox) Store(ctx context.Context, seq *proto.SequenceSet, operation proto.Operation, flags imap.FlagSet) error {
	messages, err := m.snap.getMessagesInRange(ctx, seq)
	if err != nil {
		return err
	}

	return m.state.db().Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		switch operation {
		case proto.Operation_Add:
			if err := m.state.actionAddMessageFlags(ctx, tx, messages, flags); err != nil {
				return err
			}

		case proto.Operation_Remove:
			if err := m.state.actionRemoveMessageFlags(ctx, tx, messages, flags); err != nil {
				return err
			}

		case proto.Operation_Replace:
			if err := m.state.actionSetMessageFlags(ctx, tx, messages, flags); err != nil {
				return err
			}
		}

		return nil
	})
}

func (m *Mailbox) Expunge(ctx context.Context, seq *proto.SequenceSet) error {
	var msg []*snapMsg

	if seq != nil {
		snapMsgs, err := m.snap.getMessagesInRange(ctx, seq)
		if err != nil {
			return err
		}

		msg = snapMsgs
	} else {
		msg = m.snap.getAllMessages()
	}

	return m.expunge(ctx, msg)
}

func (m *Mailbox) expunge(ctx context.Context, messages []*snapMsg) error {
	messages = xslices.Filter(messages, func(msg *snapMsg) bool {
		return msg.flags.Contains(imap.FlagDeleted)
	})

	msgIDs := xslices.Map(messages, func(msg *snapMsg) ids.MessageIDPair {
		return msg.ID
	})

	return m.state.db().Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		return m.state.actionRemoveMessagesFromMailbox(ctx, tx, msgIDs, m.snap.mboxID)
	})
}

func (m *Mailbox) Flush(ctx context.Context, permitExpunge bool) ([]response.Response, error) {
	return m.state.flushResponses(ctx, permitExpunge)
}

func (m *Mailbox) Close(ctx context.Context) error {
	m.state.deleteConnMetadata()

	return m.state.close()
}
