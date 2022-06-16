package connector

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/ticker"
	"github.com/bradenaw/juniper/xslices"
	"golang.org/x/exp/slices"
)

var (
	ErrInvalidPrefix   = errors.New("invalid prefix")
	ErrRenameForbidden = errors.New("rename operation is not allowed")
	ErrDeleteForbidden = errors.New("delete operation is not allowed")
)

type Dummy struct {
	// state holds the fake connector state.
	state *dummyState

	// usernames holds usernames that can be used for authorization.
	usernames []string

	// password holds the password that can be used for authorization.
	password string

	// These hold the default flags/attributes given to mailboxes.
	flags, permFlags, attrs imap.FlagSet

	// These hold prefixes given to folders (exclusive) and labels (non-exclusive).
	pfxFolder, pfxLabel string

	// updateCh delivers simulated updates to the mailserver.
	updateCh chan imap.Update

	// ticker controls the delivery of simulated events to the mailserver.
	ticker *ticker.Ticker

	// queue holds queued updates which are to be delivered to the mailserver each tick cycle.
	queue     []imap.Update
	queueLock sync.Mutex
}

func NewDummy(usernames []string, password string, period time.Duration, flags, permFlags, attrs imap.FlagSet) *Dummy {
	conn := &Dummy{
		state:     newDummyState(flags, permFlags, attrs),
		usernames: usernames,
		password:  password,
		flags:     flags,
		permFlags: permFlags,
		attrs:     attrs,
		updateCh:  make(chan imap.Update),
		ticker:    ticker.New(period),
	}

	go func() {
		conn.ticker.Tick(func(time.Time) {
			for _, update := range conn.popUpdates() {
				defer update.Wait()
				conn.updateCh <- update
			}
		})
	}()

	conn.state.createLabel([]string{imap.Inbox}, false)

	return conn
}

func (conn *Dummy) Authorize(username, password string) bool {
	if password != conn.password {
		return false
	}

	return slices.Contains(conn.usernames, username)
}

func (conn *Dummy) GetUpdates() <-chan imap.Update {
	return conn.updateCh
}

func (conn *Dummy) Pause() {
	conn.ticker.Pause()
}

func (conn *Dummy) Resume() {
	conn.ticker.Resume()
}

func (conn *Dummy) ValidateCreate(name []string) (imap.FlagSet, imap.FlagSet, imap.FlagSet, error) {
	if _, err := conn.validateName(name); err != nil {
		return nil, nil, nil, err
	}

	return conn.flags, conn.permFlags, conn.attrs, nil
}

func (conn *Dummy) ValidateUpdate(oldName, newName []string) error {
	oldEx, err := conn.validateName(oldName)
	if err != nil {
		return err
	}

	newEx, err := conn.validateName(newName)
	if err != nil {
		return err
	}

	if oldEx != newEx {
		return ErrRenameForbidden
	}

	return nil
}

func (conn *Dummy) ValidateDelete(name []string) error {
	if len(name) > 1 {
		return nil
	}

	if len(conn.pfxFolder) > 0 && name[0] == conn.pfxFolder {
		return ErrDeleteForbidden
	}

	if len(conn.pfxLabel) > 0 && name[0] == conn.pfxLabel {
		return ErrDeleteForbidden
	}

	return nil
}

func (conn *Dummy) GetLabel(ctx context.Context, labelID string) (imap.Mailbox, error) {
	return conn.state.getLabel(labelID), nil
}

func (conn *Dummy) CreateLabel(ctx context.Context, name []string) (imap.Mailbox, error) {
	exclusive, err := conn.validateName(name)
	if err != nil {
		return imap.Mailbox{}, nil
	}

	mbox := conn.state.createLabel(name, exclusive)

	conn.pushUpdate(imap.NewMailboxCreated(mbox))

	return mbox, nil
}

func (conn *Dummy) UpdateLabel(ctx context.Context, labelID string, name []string) error {
	conn.state.updateLabel(labelID, name)

	conn.pushUpdate(imap.NewMailboxUpdated(labelID, name))

	return nil
}

func (conn *Dummy) DeleteLabel(ctx context.Context, labelID string) error {
	conn.state.deleteLabel(labelID)

	conn.pushUpdate(imap.NewMailboxDeleted(labelID))

	return nil
}

func (conn *Dummy) GetMessage(ctx context.Context, messageID string) (imap.Message, []string, error) {
	return conn.state.getMessage(messageID), conn.state.getLabelIDs(messageID), nil
}

func (conn *Dummy) CreateMessage(ctx context.Context, mboxID string, literal []byte, flags imap.FlagSet, date time.Time) (imap.Message, error) {
	// NOTE: We are only recording this here since it was the easiest command to verify the data has been record properly
	// in the context, as APPEND will always require a communication with the remote connector.
	conn.state.recordIMAPID(ctx)

	message := conn.state.createMessage(
		mboxID,
		literal,
		flags.Contains(imap.FlagSeen),
		flags.Contains(imap.FlagFlagged),
		date,
	)

	update := imap.NewMessagesCreated()

	if err := update.Add(message, literal, []string{mboxID}); err != nil {
		return imap.Message{}, err
	}

	conn.pushUpdate(update)

	return message, nil
}

func (conn *Dummy) LabelMessages(ctx context.Context, messageIDs []string, mboxID string) error {
	for _, messageID := range messageIDs {
		conn.state.labelMessage(messageID, mboxID)

		conn.pushUpdate(imap.NewMessageUpdated(
			messageID,
			conn.state.getLabelIDs(messageID),
			conn.state.isSeen(messageID),
			conn.state.isFlagged(messageID),
		))
	}

	return nil
}

func (conn *Dummy) UnlabelMessages(ctx context.Context, messageIDs []string, mboxID string) error {
	for _, messageID := range messageIDs {
		conn.state.unlabelMessage(messageID, mboxID)

		conn.pushUpdate(imap.NewMessageUpdated(
			messageID,
			conn.state.getLabelIDs(messageID),
			conn.state.isSeen(messageID),
			conn.state.isFlagged(messageID),
		))
	}

	return nil
}

func (conn *Dummy) MarkMessagesSeen(ctx context.Context, messageIDs []string, seen bool) error {
	for _, messageID := range messageIDs {
		conn.state.setSeen(messageID, seen)

		conn.pushUpdate(imap.NewMessageUpdated(
			messageID,
			conn.state.getLabelIDs(messageID),
			conn.state.isSeen(messageID),
			conn.state.isFlagged(messageID),
		))
	}

	return nil
}

func (conn *Dummy) MarkMessagesFlagged(ctx context.Context, messageIDs []string, flagged bool) error {
	for _, messageID := range messageIDs {
		conn.state.setFlagged(messageID, flagged)

		conn.pushUpdate(imap.NewMessageUpdated(
			messageID,
			conn.state.getLabelIDs(messageID),
			conn.state.isSeen(messageID),
			conn.state.isFlagged(messageID),
		))
	}

	return nil
}

func (conn *Dummy) Sync(ctx context.Context) error {
	for _, mailbox := range conn.state.getLabels() {
		conn.updateCh <- imap.NewMailboxCreated(mailbox)
	}

	update := imap.NewMessagesCreated()

	for _, message := range conn.state.getMessages() {
		if err := update.Add(message, conn.state.getLiteral(message.ID), conn.state.getLabelIDs(message.ID)); err != nil {
			return err
		}
	}

	conn.updateCh <- update

	return nil
}

func (conn *Dummy) GetLastRecordedIMAPID() imap.ID {
	return conn.state.lastIMAPID
}

func (conn *Dummy) validateName(name []string) (bool, error) {
	var exclusive bool

	switch {
	case len(conn.pfxFolder)+len(conn.pfxLabel) == 0:
		exclusive = false

	case len(conn.pfxFolder) > 0 && len(conn.pfxLabel) > 0:
		if name[0] == conn.pfxFolder {
			exclusive = true
		} else if name[0] == conn.pfxLabel {
			exclusive = false
		} else {
			return false, ErrInvalidPrefix
		}

	case len(conn.pfxFolder) > 0:
		if len(name) > 1 && name[0] == conn.pfxFolder {
			exclusive = true
		}

	case len(conn.pfxLabel) > 0:
		if len(name) > 1 && name[0] == conn.pfxLabel {
			exclusive = false
		}
	}

	return exclusive, nil
}

func (conn *Dummy) pushUpdate(update imap.Update) {
	conn.queueLock.Lock()
	defer conn.queueLock.Unlock()

	// We mimic the behaviour of the Proton sever. if several update to a message or mailbox happen in between
	// two event polls, we only get one refresh update with the latest state.
	switch update := update.(type) {
	case *imap.MessageUpdated:
		conn.queue = removeMessageUpdatedFromSlice(conn.queue, update.MessageID)

	case *imap.MailboxUpdated:
		conn.queue = removeMailboxUpdatedFromSlice(conn.queue, update.MailboxID)
	}

	conn.queue = append(conn.queue, update)
}

func (conn *Dummy) popUpdates() []imap.Update {
	conn.queueLock.Lock()
	defer conn.queueLock.Unlock()

	var updates []imap.Update

	updates, conn.queue = conn.queue, []imap.Update{}

	return updates
}

func removeMessageUpdatedFromSlice(updates []imap.Update, messageID string) []imap.Update {
	return xslices.Filter(updates, func(update imap.Update) bool {
		u, ok := update.(*imap.MessageUpdated)

		return (!ok) || (u.MessageID != messageID)
	})
}

func removeMailboxUpdatedFromSlice(updates []imap.Update, mailboxID string) []imap.Update {
	return xslices.Filter(updates, func(update imap.Update) bool {
		u, ok := update.(*imap.MailboxUpdated)

		return (!ok) || (u.MailboxID != mailboxID)
	})
}
