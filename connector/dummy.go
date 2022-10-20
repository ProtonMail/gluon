package connector

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"time"

	"github.com/ProtonMail/gluon/constants"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/ticker"
	"github.com/bradenaw/juniper/xslices"
	"golang.org/x/exp/slices"
)

var (
	ErrNoSuchLabel   = errors.New("no such label")
	ErrNoSuchMessage = errors.New("no such message")

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
	password []byte

	// These hold the default flags/attributes given to mailboxes.
	flags, permFlags, attrs imap.FlagSet

	// These hold prefixes given to folders (exclusive) and labels (non-exclusive).
	pfxFolder, pfxLabel string

	// updateCh delivers simulated updates to the mailserver.
	updateCh     chan imap.Update
	updateQuitCh chan struct{}

	// ticker controls the delivery of simulated events to the mailserver.
	ticker *ticker.Ticker

	// queue holds queued updates which are to be delivered to the mailserver each tick cycle.
	queue     []imap.Update
	queueLock sync.Mutex

	hiddenLabels map[imap.LabelID]struct{}
}

func NewDummy(usernames []string, password []byte, period time.Duration, flags, permFlags, attrs imap.FlagSet) *Dummy {
	conn := &Dummy{
		state:        newDummyState(flags, permFlags, attrs),
		usernames:    usernames,
		password:     password,
		flags:        flags,
		permFlags:    permFlags,
		attrs:        attrs,
		updateCh:     make(chan imap.Update, constants.ChannelBufferCount),
		updateQuitCh: make(chan struct{}),
		ticker:       ticker.New(period),
		hiddenLabels: make(map[imap.LabelID]struct{}),
	}

	go func() {
		conn.ticker.Tick(func(time.Time) {
			for _, update := range conn.popUpdates() {
				defer update.Wait()

				select {
				case conn.updateCh <- update:
					continue
				case <-conn.updateQuitCh:
					return
				}
			}
		})
	}()

	conn.state.createLabelWithID([]string{imap.Inbox}, "0", false)

	return conn
}

func (conn *Dummy) Authorize(username string, password []byte) bool {
	if bytes.Compare(password, conn.password) != 0 {
		return false
	}

	return slices.Contains(conn.usernames, username)
}

func (conn *Dummy) GetUpdates() <-chan imap.Update {
	return conn.updateCh
}

func (conn *Dummy) GetLabel(ctx context.Context, labelID imap.LabelID) (imap.Mailbox, error) {
	return conn.state.getLabel(labelID)
}

func (conn *Dummy) CreateLabel(ctx context.Context, name []string) (imap.Mailbox, error) {
	exclusive, err := conn.validateName(name)
	if err != nil {
		return imap.Mailbox{}, err
	}

	mbox := conn.state.createLabel(name, exclusive)

	conn.pushUpdate(imap.NewMailboxCreated(mbox))

	return mbox, nil
}

func (conn *Dummy) UpdateLabel(ctx context.Context, labelID imap.LabelID, newName []string) error {
	label, err := conn.state.getLabel(labelID)
	if err != nil {
		return err
	}

	if err := conn.validateUpdate(label.Name, newName); err != nil {
		return err
	}

	conn.state.updateLabel(labelID, newName)

	conn.pushUpdate(imap.NewMailboxUpdated(labelID, newName))

	return nil
}

func (conn *Dummy) DeleteLabel(ctx context.Context, labelID imap.LabelID) error {
	conn.state.deleteLabel(labelID)

	conn.pushUpdate(imap.NewMailboxDeleted(labelID))

	return nil
}

func (conn *Dummy) GetMessage(ctx context.Context, messageID imap.MessageID) (imap.Message, []imap.LabelID, error) {
	message, err := conn.state.getMessage(messageID)
	if err != nil {
		return imap.Message{}, nil, err
	}

	return message, conn.state.getLabelIDs(messageID), nil
}

func (conn *Dummy) CreateMessage(ctx context.Context, mboxID imap.LabelID, literal []byte, flags imap.FlagSet, date time.Time) (imap.Message, []byte, error) {
	// NOTE: We are only recording this here since it was the easiest command to verify the data has been record properly
	// in the context, as APPEND will always require a communication with the remote connector.
	conn.state.recordIMAPID(ctx)

	parsed, err := imap.NewParsedMessage(literal)
	if err != nil {
		return imap.Message{}, nil, err
	}

	message := conn.state.createMessage(
		mboxID,
		literal,
		parsed,
		flags.ContainsUnchecked(imap.FlagSeenLowerCase),
		flags.ContainsUnchecked(imap.FlagFlaggedLowerCase),
		flags,
		date,
	)

	update := imap.NewMessagesCreated(&imap.MessageCreated{
		Message:       message,
		Literal:       literal,
		LabelIDs:      []imap.LabelID{mboxID},
		ParsedMessage: parsed,
	})

	conn.pushUpdate(update)

	return message, literal, nil
}

func (conn *Dummy) LabelMessages(ctx context.Context, messageIDs []imap.MessageID, mboxID imap.LabelID) error {
	for _, messageID := range messageIDs {
		conn.state.labelMessage(messageID, mboxID)

		conn.pushUpdate(imap.NewMessageLabelsUpdated(
			messageID,
			conn.state.getLabelIDs(messageID),
			conn.state.isSeen(messageID),
			conn.state.isFlagged(messageID),
		))
	}

	return nil
}

func (conn *Dummy) UnlabelMessages(ctx context.Context, messageIDs []imap.MessageID, mboxID imap.LabelID) error {
	for _, messageID := range messageIDs {
		conn.state.unlabelMessage(messageID, mboxID)

		conn.pushUpdate(imap.NewMessageLabelsUpdated(
			messageID,
			conn.state.getLabelIDs(messageID),
			conn.state.isSeen(messageID),
			conn.state.isFlagged(messageID),
		))
	}

	return nil
}

func (conn *Dummy) MoveMessages(ctx context.Context, messageIDs []imap.MessageID, labelFromID, labelToID imap.LabelID) error {
	for _, messageID := range messageIDs {
		conn.state.unlabelMessage(messageID, labelFromID)
		conn.state.labelMessage(messageID, labelToID)

		conn.pushUpdate(imap.NewMessageLabelsUpdated(
			messageID,
			conn.state.getLabelIDs(messageID),
			conn.state.isSeen(messageID),
			conn.state.isFlagged(messageID),
		))
	}

	return nil
}

func (conn *Dummy) MarkMessagesSeen(ctx context.Context, messageIDs []imap.MessageID, seen bool) error {
	for _, messageID := range messageIDs {
		conn.state.setSeen(messageID, seen)

		conn.pushUpdate(imap.NewMessageFlagsUpdated(
			messageID,
			conn.state.isSeen(messageID),
			conn.state.isFlagged(messageID),
		))
	}

	return nil
}

func (conn *Dummy) MarkMessagesFlagged(ctx context.Context, messageIDs []imap.MessageID, flagged bool) error {
	for _, messageID := range messageIDs {
		conn.state.setFlagged(messageID, flagged)

		conn.pushUpdate(imap.NewMessageFlagsUpdated(
			messageID,
			conn.state.isSeen(messageID),
			conn.state.isFlagged(messageID),
		))
	}

	return nil
}

func (conn *Dummy) GetUIDValidity() imap.UID {
	return 1
}

func (conn *Dummy) SetUIDValidity(imap.UID) error {
	return nil
}

func (conn *Dummy) Sync(ctx context.Context) error {
	for _, mailbox := range conn.state.getLabels() {
		update := imap.NewMailboxCreated(mailbox)
		defer update.WaitContext(ctx)

		conn.updateCh <- update
	}

	var updates []*imap.MessageCreated

	for _, message := range conn.state.getMessages() {
		update, err := conn.state.getMessageCreatedUpdate(message.ID)
		if err != nil {
			return err
		}

		updates = append(updates, update)
	}

	update := imap.NewMessagesCreated(updates...)
	defer update.WaitContext(ctx)

	conn.updateCh <- update

	return nil
}

func (conn *Dummy) Close(ctx context.Context) error {
	close(conn.updateQuitCh)
	close(conn.updateCh)
	conn.ticker.Stop()
	conn.password = nil

	return nil
}

func (conn *Dummy) GetLastRecordedIMAPID() imap.IMAPID {
	return conn.state.lastIMAPID
}

func (conn *Dummy) ClearUpdates() {
	conn.popUpdates()
}

func (conn *Dummy) IsLabelVisible(_ context.Context, id imap.LabelID) bool {
	_, ok := conn.hiddenLabels[id]

	return !ok
}

func (conn *Dummy) SetMailboxVisible(id imap.LabelID, visible bool) {
	if !visible {
		conn.hiddenLabels[id] = struct{}{}
	} else {
		delete(conn.hiddenLabels, id)
	}
}

func (conn *Dummy) pushUpdate(update imap.Update) {
	conn.queueLock.Lock()
	defer conn.queueLock.Unlock()

	// We mimic the behaviour of the Proton sever. if several update to a message or mailbox happen in between
	// two event polls, we only get one refresh update with the latest state.
	switch update := update.(type) {
	case *imap.MessageLabelsUpdated:
		conn.queue = removeMessageLabelsUpdatedFromSlice(conn.queue, update.MessageID)
	case *imap.MessageFlagsUpdated:
		conn.queue = removeMessageFlagsUpdatedFromSlice(conn.queue, update.MessageID)

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

func (conn *Dummy) validateUpdate(oldName, newName []string) error {
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

func removeMessageLabelsUpdatedFromSlice(updates []imap.Update, messageID imap.MessageID) []imap.Update {
	return xslices.Filter(updates, func(update imap.Update) bool {
		u, ok := update.(*imap.MessageLabelsUpdated)

		return (!ok) || (u.MessageID != messageID)
	})
}

func removeMessageFlagsUpdatedFromSlice(updates []imap.Update, messageID imap.MessageID) []imap.Update {
	return xslices.Filter(updates, func(update imap.Update) bool {
		u, ok := update.(*imap.MessageFlagsUpdated)

		return (!ok) || (u.MessageID != messageID)
	})
}

func removeMailboxUpdatedFromSlice(updates []imap.Update, mailboxID imap.LabelID) []imap.Update {
	return xslices.Filter(updates, func(update imap.Update) bool {
		u, ok := update.(*imap.MailboxUpdated)

		return (!ok) || (u.MailboxID != mailboxID)
	})
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
