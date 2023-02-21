package connector

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ProtonMail/gluon/constants"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/ticker"
	"github.com/bradenaw/juniper/xslices"
	"golang.org/x/exp/slices"
)

var (
	ErrNoSuchMailbox = errors.New("no such mailbox")
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

	// hiddenMailboxes holds mailboxes that are hidden from the user.
	hiddenMailboxes map[imap.MailboxID]struct{}

	// uidValidity holds the global UID validity.
	uidValidity imap.UID

	allowMessageCreateWithUnknownMailboxID bool

	updatesAllowedToFail int32
}

func NewDummy(usernames []string, password []byte, period time.Duration, flags, permFlags, attrs imap.FlagSet) *Dummy {
	conn := &Dummy{
		state:           newDummyState(flags, permFlags, attrs),
		usernames:       usernames,
		password:        password,
		flags:           flags,
		permFlags:       permFlags,
		attrs:           attrs,
		updateCh:        make(chan imap.Update, constants.ChannelBufferCount),
		updateQuitCh:    make(chan struct{}),
		ticker:          ticker.New(period),
		hiddenMailboxes: make(map[imap.MailboxID]struct{}),
		uidValidity:     1,
	}

	go func() {
		conn.ticker.Tick(func(time.Time) {
			for _, update := range conn.popUpdates() {
				defer func() {
					err, ok := update.Wait()
					if ok && err != nil {
						if atomic.LoadInt32(&conn.updatesAllowedToFail) == 0 {
							panic(fmt.Sprintf("Failed to apply update %v: %v", update.String(), err))
						} else {
							logrus.Errorf("Failed to apply update %v: %v", update.String(), err)
						}
					}
				}()

				select {
				case conn.updateCh <- update:
					continue
				case <-conn.updateQuitCh:
					return
				}
			}
		})
	}()

	conn.state.createMailboxWithID([]string{imap.Inbox}, "0", false)

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

func (conn *Dummy) CreateMailbox(ctx context.Context, name []string) (imap.Mailbox, error) {
	exclusive, err := conn.validateName(name)
	if err != nil {
		return imap.Mailbox{}, err
	}

	mbox := conn.state.createMailbox(name, exclusive)

	conn.pushUpdate(imap.NewMailboxCreated(mbox))

	return mbox, nil
}

func (conn *Dummy) UpdateMailboxName(ctx context.Context, mboxID imap.MailboxID, newName []string) error {
	mbox, err := conn.state.getMailbox(mboxID)
	if err != nil {
		return err
	}

	if err := conn.validateUpdate(mbox.Name, newName); err != nil {
		return err
	}

	conn.state.updateMailboxName(mboxID, newName)

	conn.pushUpdate(imap.NewMailboxUpdated(mboxID, newName))

	return nil
}

func (conn *Dummy) DeleteMailbox(ctx context.Context, mboxID imap.MailboxID) error {
	conn.state.deleteMailbox(mboxID)

	conn.pushUpdate(imap.NewMailboxDeleted(mboxID))

	return nil
}

func (conn *Dummy) GetMessageLiteral(ctx context.Context, id imap.MessageID) ([]byte, error) {
	return conn.state.tryGetLiteral(id)
}

func (conn *Dummy) CreateMessage(ctx context.Context, mboxID imap.MailboxID, literal []byte, flags imap.FlagSet, date time.Time) (imap.Message, []byte, error) {
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

	update := imap.NewMessagesCreated(conn.allowMessageCreateWithUnknownMailboxID, &imap.MessageCreated{
		Message:       message,
		Literal:       literal,
		MailboxIDs:    []imap.MailboxID{mboxID},
		ParsedMessage: parsed,
	})

	conn.pushUpdate(update)

	return message, literal, nil
}

func (conn *Dummy) AddMessagesToMailbox(ctx context.Context, messageIDs []imap.MessageID, mboxID imap.MailboxID) error {
	for _, messageID := range messageIDs {
		conn.state.addMessageToMailbox(messageID, mboxID)

		conn.pushUpdate(imap.NewMessageMailboxesUpdated(
			messageID,
			conn.state.getMailboxIDs(messageID),
			conn.state.isSeen(messageID),
			conn.state.isFlagged(messageID),
		))
	}

	return nil
}

func (conn *Dummy) RemoveMessagesFromMailbox(ctx context.Context, messageIDs []imap.MessageID, mboxID imap.MailboxID) error {
	for _, messageID := range messageIDs {
		conn.state.removeMessageFromMailbox(messageID, mboxID)

		conn.pushUpdate(imap.NewMessageMailboxesUpdated(
			messageID,
			conn.state.getMailboxIDs(messageID),
			conn.state.isSeen(messageID),
			conn.state.isFlagged(messageID),
		))
	}

	return nil
}

func (conn *Dummy) MoveMessages(ctx context.Context, messageIDs []imap.MessageID, mboxFromID, mboxToID imap.MailboxID) (bool, error) {
	for _, messageID := range messageIDs {
		conn.state.removeMessageFromMailbox(messageID, mboxFromID)
		conn.state.addMessageToMailbox(messageID, mboxToID)

		conn.pushUpdate(imap.NewMessageMailboxesUpdated(
			messageID,
			conn.state.getMailboxIDs(messageID),
			conn.state.isSeen(messageID),
			conn.state.isFlagged(messageID),
		))
	}

	return true, nil
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
	return conn.uidValidity
}

func (conn *Dummy) SetUIDValidity(newUIDValidity imap.UID) error {
	conn.uidValidity = newUIDValidity

	return nil
}

func (conn *Dummy) Sync(ctx context.Context) error {
	for _, mailbox := range conn.state.getMailboxes() {
		update := imap.NewMailboxCreated(mailbox)

		conn.updateCh <- update

		err, ok := update.WaitContext(ctx)
		if ok && err != nil {
			return fmt.Errorf("failed to apply update %v:%w", update.String(), err)
		}
	}

	var updates []*imap.MessageCreated

	for _, message := range conn.state.getMessages() {
		update, err := conn.state.getMessageCreatedUpdate(message.ID)
		if err != nil {
			return err
		}

		updates = append(updates, update)
	}

	update := imap.NewMessagesCreated(conn.allowMessageCreateWithUnknownMailboxID, updates...)

	conn.updateCh <- update

	err, ok := update.WaitContext(ctx)
	if ok && err != nil {
		return fmt.Errorf("failed to apply update %v:%w", update.String(), err)
	}

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

func (conn *Dummy) IsMailboxVisible(_ context.Context, id imap.MailboxID) bool {
	_, ok := conn.hiddenMailboxes[id]

	return !ok
}

func (conn *Dummy) SetMailboxVisible(id imap.MailboxID, visible bool) {
	if !visible {
		conn.hiddenMailboxes[id] = struct{}{}
	} else {
		delete(conn.hiddenMailboxes, id)
	}
}

func (conn *Dummy) pushUpdate(update imap.Update) {
	conn.queueLock.Lock()
	defer conn.queueLock.Unlock()

	// We mimic the behaviour of the Proton sever. if several update to a message or mailbox happen in between
	// two event polls, we only get one refresh update with the latest state.
	switch update := update.(type) {
	case *imap.MessageMailboxesUpdated:
		conn.queue = removeMessageMailboxesUpdatedFromSlice(conn.queue, update.MessageID)

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

func removeMessageMailboxesUpdatedFromSlice(updates []imap.Update, messageID imap.MessageID) []imap.Update {
	return xslices.Filter(updates, func(update imap.Update) bool {
		u, ok := update.(*imap.MessageMailboxesUpdated)

		return (!ok) || (u.MessageID != messageID)
	})
}

func removeMessageFlagsUpdatedFromSlice(updates []imap.Update, messageID imap.MessageID) []imap.Update {
	return xslices.Filter(updates, func(update imap.Update) bool {
		u, ok := update.(*imap.MessageFlagsUpdated)

		return (!ok) || (u.MessageID != messageID)
	})
}

func removeMailboxUpdatedFromSlice(updates []imap.Update, mailboxID imap.MailboxID) []imap.Update {
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

func (conn *Dummy) SetUpdatesAllowedToFail(value bool) {
	var v int32
	if value {
		v = 1
	} else {
		v = 0
	}

	atomic.StoreInt32(&conn.updatesAllowedToFail, v)
}
