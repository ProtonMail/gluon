package state

import (
	"context"
	"strings"
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db"
	"github.com/ProtonMail/gluon/internal/db/ent"
	"github.com/ProtonMail/gluon/internal/ids"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/gluon/rfc822"
	"github.com/bradenaw/juniper/xslices"
	"github.com/sirupsen/logrus"
)

type Mailbox struct {
	id          ids.MailboxIDPair
	name        string
	subscribed  bool
	uidValidity imap.UID
	uidNext     imap.UID

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
		id:          ids.NewMailboxIDPair(mbox),
		name:        mbox.Name,
		uidValidity: mbox.UIDValidity,
		uidNext:     mbox.UIDNext,

		state: state,

		selected: snap != nil,
		readOnly: state.ro,
		snap:     snap,
	}
}

func (m *Mailbox) Name() string {
	return m.name
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
	return m.snap.len()
}

func (m *Mailbox) Flags(ctx context.Context) (imap.FlagSet, error) {
	return db.ReadResult(ctx, m.state.db(), func(ctx context.Context, client *ent.Client) (imap.FlagSet, error) {
		return db.GetMailboxFlags(ctx, client, m.id.InternalID)
	})
}

func (m *Mailbox) PermanentFlags(ctx context.Context) (imap.FlagSet, error) {
	return db.ReadResult(ctx, m.state.db(), func(ctx context.Context, client *ent.Client) (imap.FlagSet, error) {
		return db.GetMailboxPermanentFlags(ctx, client, m.id.InternalID)
	})
}

func (m *Mailbox) Attributes(ctx context.Context) (imap.FlagSet, error) {
	return db.ReadResult(ctx, m.state.db(), func(ctx context.Context, client *ent.Client) (imap.FlagSet, error) {
		return db.GetMailboxAttributes(ctx, client, m.id.InternalID)
	})
}

func (m *Mailbox) UIDNext() imap.UID {
	return m.uidNext
}

func (m *Mailbox) UIDValidity() imap.UID {
	return m.uidValidity
}

func (m *Mailbox) Subscribed() bool {
	return m.subscribed
}

func (m *Mailbox) GetMessagesWithFlag(flag string) []imap.SeqID {
	return xslices.Map(m.snap.getMessagesWithFlag(flag), func(msg snapMsgWithSeq) imap.SeqID {
		return msg.Seq
	})
}

func (m *Mailbox) GetFirstMessageWithFlag(flag string) (snapMsgWithSeq, bool) {
	msg, ok := m.snap.firstMessageWithFlag(flag)

	return msg, ok
}

func (m *Mailbox) GetFirstMessageWithoutFlag(flag string) (snapMsgWithSeq, bool) {
	msg, ok := m.snap.firstMessageWithoutFlag(flag)

	return msg, ok
}

func (m *Mailbox) GetMessagesWithFlagCount(flag string) int {
	return m.snap.getMessagesWithFlagCount(flag)
}

func (m *Mailbox) GetMessagesWithoutFlag(flag string) []imap.SeqID {
	return xslices.Map(m.snap.getMessagesWithoutFlag(flag), func(msg snapMsgWithSeq) imap.SeqID {
		return msg.Seq
	})
}

func (m *Mailbox) GetMessagesWithoutFlagCount(flag string) int {
	return m.snap.getMessagesWithoutFlagCount(flag)
}

func (m *Mailbox) AppendRegular(ctx context.Context, literal []byte, flags imap.FlagSet, date time.Time) (imap.UID, error) {
	if err := m.state.db().Read(ctx, func(ctx context.Context, client *ent.Client) error {
		if messageCount, uid, err := db.GetMailboxMessageCountAndUID(ctx, client, m.snap.mboxID.InternalID); err != nil {
			return err
		} else {
			if err := m.state.imapLimits.CheckMailBoxMessageCount(messageCount, 1); err != nil {
				return err
			}

			if err := m.state.imapLimits.CheckUIDCount(uid, 1); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return 0, err
	}

	var appendIntoDrafts bool

	attr, err := m.Attributes(ctx)
	if err != nil {
		return 0, err
	}

	// Force create message when appending to drafts so that IMAP clients can create new draft messages.
	if !attr.Contains(imap.AttrDrafts) {
		internalIDString, err := rfc822.GetHeaderValue(literal, ids.InternalIDKey)
		if err != nil {
			return 0, err
		}

		if len(internalIDString) > 0 {
			msgID, err := imap.InternalMessageIDFromString(internalIDString)
			if err != nil {
				return 0, err
			}

			if message, err := db.ReadResult(ctx, m.state.db(), func(ctx context.Context, client *ent.Client) (*ent.Message, error) {
				message, err := db.GetMessageWithIDWithDeletedFlag(ctx, client, msgID)
				if err != nil {
					if ent.IsNotFound(err) {
						return nil, nil
					}

					return nil, err
				}

				return message, nil
			}); err != nil || message == nil {
				logrus.WithError(err).Warn("The message has an unknown internal ID")
			} else if !message.Deleted {
				logrus.Debugf("Appending duplicate message with Internal ID:%v", msgID.ShortID())
				// Only shuffle around messages that haven't been marked for deletion.
				if res, err := db.WriteResult(ctx, m.state.db(), func(ctx context.Context, tx *ent.Tx) ([]db.UIDWithFlags, error) {
					remoteID, err := db.GetMessageRemoteIDFromID(ctx, tx.Client(), msgID)
					if err != nil {
						return nil, err
					}

					return m.state.actionAddMessagesToMailbox(ctx, tx,
						[]ids.MessageIDPair{{InternalID: msgID, RemoteID: remoteID}},
						m.id,
						m.snap == m.state.snap,
					)
				}); err != nil {
					return 0, err
				} else {
					return res[0].UID, nil
				}
			}
		}
	} else {
		appendIntoDrafts = true
		newLiteral, err := rfc822.EraseHeaderValue(literal, ids.InternalIDKey)
		if err != nil {
			logrus.WithError(err).Error("Failed to erase Gluon internal id from draft")
		} else {
			literal = newLiteral
		}
	}

	return db.WriteResult(ctx, m.state.db(), func(ctx context.Context, tx *ent.Tx) (imap.UID, error) {
		return m.state.actionCreateMessage(ctx, tx, m.snap.mboxID, literal, flags, date, m.snap == m.state.snap, appendIntoDrafts)
	})
}

func (m *Mailbox) Append(ctx context.Context, literal []byte, flags imap.FlagSet, date time.Time) (imap.UID, error) {
	uid, err := m.AppendRegular(ctx, literal, flags, date)
	if err != nil {
		// Failed to append to mailbox attempt to insert into recovery mailbox.
		if err := m.state.db().Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
			return m.state.actionCreateRecoveredMessage(ctx, tx, literal, flags, date)
		}); err != nil {
			logrus.WithError(err).Error("Failed to insert message into recovery mailbox")
			reporter.ExceptionWithContext(ctx, "Failed to insert message into recovery mailbox", reporter.Context{"error": err})
		}
	}

	return uid, err
}

// Copy copies the messages represented by the given sequence set into the mailbox with the given name.
// If the context is a UID context, the sequence set refers to message UIDs.
// If no items are copied the response object will be nil.
func (m *Mailbox) Copy(ctx context.Context, seq *proto.SequenceSet, name string) (response.Item, error) {
	if strings.EqualFold(name, ids.GluonRecoveryMailboxName) {
		return nil, ErrOperationNotAllowed
	}

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
		if m.state.user.GetRecoveryMailboxID().InternalID == m.snap.mboxID.InternalID {
			return m.state.actionCopyMessagesOutOfRecoveryMailbox(ctx, tx, msgIDs, ids.NewMailboxIDPair(mbox))
		} else {
			return m.state.actionAddMessagesToMailbox(ctx, tx, msgIDs, ids.NewMailboxIDPair(mbox), m.snap == m.state.snap)
		}
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
	if strings.EqualFold(name, ids.GluonRecoveryMailboxName) {
		return nil, ErrOperationNotAllowed
	}

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
		if m.state.user.GetRecoveryMailboxID().InternalID == m.snap.mboxID.InternalID {
			return m.state.actionMoveMessagesOutOfRecoveryMailbox(ctx, tx, msgIDs, ids.NewMailboxIDPair(mbox))
		} else {
			return m.state.actionMoveMessages(ctx, tx, msgIDs, m.snap.mboxID, ids.NewMailboxIDPair(mbox))
		}
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
	var msgIDs []ids.MessageIDPair

	if seq != nil {
		snapMsgs, err := m.snap.getMessagesInRange(ctx, seq)
		if err != nil {
			return err
		}

		msgIDs = make([]ids.MessageIDPair, 0, len(snapMsgs))

		for _, v := range snapMsgs {
			if v.toExpunge {
				msgIDs = append(msgIDs, v.ID)
			}
		}
	} else {
		msgIDs = m.snap.getAllMessagesIDsMarkedDelete()
	}

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
