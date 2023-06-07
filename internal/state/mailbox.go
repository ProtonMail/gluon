package state

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ProtonMail/gluon/connector"
	"github.com/ProtonMail/gluon/db"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/imap/command"
	"github.com/ProtonMail/gluon/internal/ids"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/gluon/rfc822"
	"github.com/bradenaw/juniper/xslices"
	"github.com/sirupsen/logrus"
)

type Mailbox struct {
	id          db.MailboxIDPair
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
	IsDrafts(ctx context.Context) (bool, error)
}

func newMailbox(mbox *db.Mailbox, state *State, snap *snapshot) *Mailbox {
	return &Mailbox{
		id:          db.NewMailboxIDPair(mbox),
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
	return stateDBReadResult(ctx, m.state, func(ctx context.Context, client db.ReadOnly) (imap.FlagSet, error) {
		return client.GetMailboxFlags(ctx, m.id.InternalID)
	})
}

func (m *Mailbox) PermanentFlags(ctx context.Context) (imap.FlagSet, error) {
	return stateDBReadResult(ctx, m.state, func(ctx context.Context, client db.ReadOnly) (imap.FlagSet, error) {
		return client.GetMailboxPermanentFlags(ctx, m.id.InternalID)
	})
}

func (m *Mailbox) Attributes(ctx context.Context) (imap.FlagSet, error) {
	return stateDBReadResult(ctx, m.state, func(ctx context.Context, client db.ReadOnly) (imap.FlagSet, error) {
		return client.GetMailboxAttributes(ctx, m.id.InternalID)
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
	if err := stateDBRead(ctx, m.state, func(ctx context.Context, client db.ReadOnly) error {
		if messageCount, uid, err := client.GetMailboxMessageCountAndUID(ctx, m.snap.mboxID.InternalID); err != nil {
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

			if messageDeleted, err := stateDBReadResult(ctx, m.state, func(ctx context.Context, client db.ReadOnly) (bool, error) {
				return client.GetMessageDeletedFlag(ctx, msgID)
			}); err != nil {
				logrus.WithError(err).Warn("The message has an unknown internal ID")
			} else if !messageDeleted {
				logrus.Debugf("Appending duplicate message with Internal ID:%v", msgID.ShortID())
				// Only shuffle around messages that haven't been marked for deletion.
				if res, err := stateDBWriteResult(ctx, m.state, func(ctx context.Context, tx db.Transaction) ([]Update, []db.UIDWithFlags, error) {
					remoteID, err := tx.GetMessageRemoteID(ctx, msgID)
					if err != nil {
						return nil, nil, err
					}

					return m.state.actionAddMessagesToMailbox(ctx, tx,
						[]db.MessageIDPair{{InternalID: msgID, RemoteID: remoteID}},
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

	return stateDBWriteResult(ctx, m.state, func(ctx context.Context, tx db.Transaction) ([]Update, imap.UID, error) {
		return m.state.actionCreateMessage(ctx, tx, m.snap.mboxID, literal, flags, date, m.snap == m.state.snap, appendIntoDrafts)
	})
}

var ErrKnownRecoveredMessage = errors.New("known recovered message, possible duplication")

func (m *Mailbox) Append(ctx context.Context, literal []byte, flags imap.FlagSet, date time.Time) (imap.UID, error) {
	uid, err := m.AppendRegular(ctx, literal, flags, date)
	if err != nil {
		// Can't store messages that exceed size limits
		if errors.Is(err, connector.ErrMessageSizeExceedsLimits) {
			return uid, err
		}

		// Failed to append to mailbox attempt to insert into recovery mailbox.
		knownMessage, recoverErr := stateDBWriteResult(ctx, m.state, func(ctx context.Context, tx db.Transaction) ([]Update, bool, error) {
			return m.state.actionCreateRecoveredMessage(ctx, tx, literal, flags, date)
		})
		if recoverErr != nil && !knownMessage {
			logrus.WithError(recoverErr).Error("Failed to insert message into recovery mailbox")
			reporter.ExceptionWithContext(ctx, "Failed to insert message into recovery mailbox", reporter.Context{"error": recoverErr})
		}

		if knownMessage {
			err = fmt.Errorf("%v: %w", err, ErrKnownRecoveredMessage)
		}
	}

	return uid, err
}

func (m *Mailbox) IsDrafts(ctx context.Context) (bool, error) {
	attrs, err := m.Attributes(ctx)
	if err != nil {
		return false, err
	}

	return attrs.Contains(imap.AttrDrafts), nil
}

// Copy copies the messages represented by the given sequence set into the mailbox with the given name.
// If the context is a UID context, the sequence set refers to message UIDs.
// If no items are copied the response object will be nil.
func (m *Mailbox) Copy(ctx context.Context, seq []command.SeqRange, name string) (response.Item, error) {
	if strings.EqualFold(name, ids.GluonRecoveryMailboxName) {
		return nil, ErrOperationNotAllowed
	}

	mbox, err := stateDBReadResult(ctx, m.state, func(ctx context.Context, client db.ReadOnly) (*db.Mailbox, error) {
		return client.GetMailboxByName(ctx, name)
	})
	if err != nil {
		return nil, ErrNoSuchMailbox
	}

	messages, err := m.snap.getMessagesInRange(ctx, seq)
	if err != nil {
		return nil, err
	}

	msgIDs := make([]db.MessageIDPair, len(messages))
	msgUIDs := make([]imap.UID, len(messages))

	for i := 0; i < len(messages); i++ {
		snapMsg := messages[i]
		msgUIDs[i] = snapMsg.UID
		msgIDs[i] = snapMsg.ID
	}

	destUIDs, err := stateDBWriteResult(ctx, m.state, func(ctx context.Context, tx db.Transaction) ([]Update, []db.UIDWithFlags, error) {
		if m.state.user.GetRecoveryMailboxID().InternalID == m.snap.mboxID.InternalID {
			return m.state.actionCopyMessagesOutOfRecoveryMailbox(ctx, tx, msgIDs, db.NewMailboxIDPair(mbox))
		} else {
			return m.state.actionAddMessagesToMailbox(ctx, tx, msgIDs, db.NewMailboxIDPair(mbox), m.snap == m.state.snap)
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
func (m *Mailbox) Move(ctx context.Context, seq []command.SeqRange, name string) (response.Item, error) {
	if strings.EqualFold(name, ids.GluonRecoveryMailboxName) {
		return nil, ErrOperationNotAllowed
	}

	mbox, err := stateDBReadResult(ctx, m.state, func(ctx context.Context, client db.ReadOnly) (*db.Mailbox, error) {
		return client.GetMailboxByName(ctx, name)
	})
	if err != nil {
		return nil, ErrNoSuchMailbox
	}

	messages, err := m.snap.getMessagesInRange(ctx, seq)
	if err != nil {
		return nil, err
	}

	msgIDs := make([]db.MessageIDPair, len(messages))
	msgUIDs := make([]imap.UID, len(messages))

	for i := 0; i < len(messages); i++ {
		snapMsg := messages[i]
		msgUIDs[i] = snapMsg.UID
		msgIDs[i] = snapMsg.ID
	}

	destUIDs, err := stateDBWriteResult(ctx, m.state, func(ctx context.Context, tx db.Transaction) ([]Update, []db.UIDWithFlags, error) {
		if m.state.user.GetRecoveryMailboxID().InternalID == m.snap.mboxID.InternalID {
			return m.state.actionMoveMessagesOutOfRecoveryMailbox(ctx, tx, msgIDs, db.NewMailboxIDPair(mbox))
		} else {
			return m.state.actionMoveMessages(ctx, tx, msgIDs, m.snap.mboxID, db.NewMailboxIDPair(mbox))
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

func (m *Mailbox) Store(ctx context.Context, seqSet []command.SeqRange, action command.StoreAction, flags imap.FlagSet) error {
	messages, err := m.snap.getMessagesInRange(ctx, seqSet)
	if err != nil {
		return err
	}

	return stateDBWrite(ctx, m.state, func(ctx context.Context, tx db.Transaction) ([]Update, error) {
		switch action {
		case command.StoreActionAddFlags:
			return m.state.actionAddMessageFlags(ctx, tx, messages, flags)

		case command.StoreActionRemFlags:
			return m.state.actionRemoveMessageFlags(ctx, tx, messages, flags)

		case command.StoreActionSetFlags:
			return m.state.actionSetMessageFlags(ctx, tx, messages, flags)
		}

		return nil, fmt.Errorf("unknown flag action")
	})
}

func (m *Mailbox) Expunge(ctx context.Context, seq []command.SeqRange) error {
	var msgIDs []db.MessageIDPair

	if seq != nil {
		snapMsgs, err := m.snap.getMessagesInRange(ctx, seq)
		if err != nil {
			return err
		}

		msgIDs = make([]db.MessageIDPair, 0, len(snapMsgs))

		for _, v := range snapMsgs {
			if v.toExpunge {
				msgIDs = append(msgIDs, v.ID)
			}
		}
	} else {
		msgIDs = m.snap.getAllMessagesIDsMarkedDelete()
	}

	return stateDBWrite(ctx, m.state, func(ctx context.Context, tx db.Transaction) ([]Update, error) {
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
