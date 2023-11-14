package connector

import (
	"context"
	"sync"
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/bradenaw/juniper/xslices"
	"github.com/google/uuid"
	"golang.org/x/exp/maps"
)

type dummyState struct {
	flags, permFlags, attrs imap.FlagSet

	messages   map[imap.MessageID]*dummyMessage
	mailboxes  map[imap.MailboxID]*dummyMailbox
	lastIMAPID imap.IMAPID

	lock sync.RWMutex
}

type dummyMailbox struct {
	mboxName  []string
	exclusive bool
}

type dummyMessage struct {
	literal   []byte
	parsed    *imap.ParsedMessage
	seen      bool
	flagged   bool
	forwarded bool
	date      time.Time
	flags     imap.FlagSet

	mboxIDs map[imap.MailboxID]struct{}
}

func newDummyState(flags, permFlags, attrs imap.FlagSet) *dummyState {
	return &dummyState{
		flags:      flags,
		permFlags:  permFlags,
		attrs:      attrs,
		messages:   make(map[imap.MessageID]*dummyMessage),
		mailboxes:  make(map[imap.MailboxID]*dummyMailbox),
		lastIMAPID: imap.NewIMAPID(),
	}
}

func (state *dummyState) recordIMAPID(ctx context.Context) {
	if id, ok := imap.GetIMAPIDFromContext(ctx); ok {
		state.lock.Lock()
		defer state.lock.Unlock()
		state.lastIMAPID = id
	}
}

func (state *dummyState) getMailboxes() []imap.Mailbox {
	state.lock.Lock()
	defer state.lock.Unlock()

	return xslices.Map(maps.Keys(state.mailboxes), func(mboxID imap.MailboxID) imap.Mailbox {
		return state.toMailbox(mboxID)
	})
}

func (state *dummyState) getMailbox(mboxID imap.MailboxID) (imap.Mailbox, error) {
	state.lock.Lock()
	defer state.lock.Unlock()

	if _, ok := state.mailboxes[mboxID]; !ok {
		return imap.Mailbox{}, ErrNoSuchMailbox
	}

	return state.toMailbox(mboxID), nil
}

func (state *dummyState) createMailbox(name []string, exclusive bool) imap.Mailbox {
	state.lock.Lock()
	defer state.lock.Unlock()

	mboxID := imap.MailboxID(uuid.NewString())

	state.mailboxes[mboxID] = &dummyMailbox{
		mboxName:  name,
		exclusive: exclusive,
	}

	return state.toMailbox(mboxID)
}

func (state *dummyState) createMailboxWithID(name []string, id imap.MailboxID, exclusive bool) imap.Mailbox {
	state.lock.Lock()
	defer state.lock.Unlock()

	state.mailboxes[id] = &dummyMailbox{
		mboxName:  name,
		exclusive: exclusive,
	}

	return state.toMailbox(id)
}

func (state *dummyState) updateMailboxName(mboxID imap.MailboxID, name []string) {
	state.lock.Lock()
	defer state.lock.Unlock()

	state.mailboxes[mboxID].mboxName = name
}

func (state *dummyState) renameMailbox(mboxID imap.MailboxID, name []string) {
	state.lock.Lock()
	defer state.lock.Unlock()

	state.mailboxes[mboxID].mboxName = name
}

func (state *dummyState) deleteMailbox(mboxID imap.MailboxID) {
	state.lock.Lock()
	defer state.lock.Unlock()

	delete(state.mailboxes, mboxID)

	for _, message := range state.messages {
		delete(message.mboxIDs, mboxID)
	}
}

func (state *dummyState) getMessages() []imap.Message {
	state.lock.Lock()
	defer state.lock.Unlock()

	return xslices.Map(maps.Keys(state.messages), func(messageID imap.MessageID) imap.Message {
		return state.toMessage(messageID)
	})
}

func (state *dummyState) getMessageCreatedUpdate(id imap.MessageID) (*imap.MessageCreated, error) {
	state.lock.Lock()
	defer state.lock.Unlock()

	msg, ok := state.messages[id]
	if !ok {
		return nil, ErrNoSuchMessage
	}

	return &imap.MessageCreated{
		Message:       state.toMessage(id),
		Literal:       msg.literal,
		MailboxIDs:    maps.Keys(msg.mboxIDs),
		ParsedMessage: msg.parsed,
	}, nil
}

func (state *dummyState) getMessage(messageID imap.MessageID) (imap.Message, error) {
	state.lock.Lock()
	defer state.lock.Unlock()

	if _, ok := state.messages[messageID]; !ok {
		return imap.Message{}, ErrNoSuchMessage
	}

	return state.toMessage(messageID), nil
}

func (state *dummyState) getMailboxIDs(messageID imap.MessageID) []imap.MailboxID {
	state.lock.Lock()
	defer state.lock.Unlock()

	return maps.Keys(state.messages[messageID].mboxIDs)
}

func (state *dummyState) getLiteral(messageID imap.MessageID) []byte {
	state.lock.Lock()
	defer state.lock.Unlock()

	return state.messages[messageID].literal
}

func (state *dummyState) tryGetLiteral(messageID imap.MessageID) ([]byte, error) {
	state.lock.Lock()
	defer state.lock.Unlock()

	v, ok := state.messages[messageID]
	if !ok {
		return nil, ErrNoSuchMessage
	}

	return v.literal, nil
}

func (state *dummyState) createMessage(
	mboxID imap.MailboxID,
	literal []byte,
	parsed *imap.ParsedMessage,
	seen, flagged bool,
	otherFlags imap.FlagSet,
	date time.Time,
) imap.Message {
	state.lock.Lock()
	defer state.lock.Unlock()

	messageID := imap.MessageID(uuid.NewString())

	if seen {
		otherFlags.RemoveFromSelf(imap.FlagSeen)
	}

	if flagged {
		otherFlags.RemoveFromSelf(imap.FlagFlagged)
	}

	state.messages[messageID] = &dummyMessage{
		literal: literal,
		seen:    seen,
		parsed:  parsed,
		flagged: flagged,
		flags:   otherFlags,
		date:    date,
		mboxIDs: map[imap.MailboxID]struct{}{mboxID: {}},
	}

	return state.toMessage(messageID)
}

func (state *dummyState) addMessageToMailbox(messageID imap.MessageID, mboxID imap.MailboxID) {
	state.lock.Lock()
	defer state.lock.Unlock()

	if state.mailboxes[mboxID].exclusive {
		state.messages[messageID].mboxIDs = make(map[imap.MailboxID]struct{})
	}

	state.messages[messageID].mboxIDs[mboxID] = struct{}{}
}

func (state *dummyState) removeMessageFromMailbox(messageID imap.MessageID, mboxID imap.MailboxID) {
	state.lock.Lock()
	defer state.lock.Unlock()

	delete(state.messages[messageID].mboxIDs, mboxID)
}

func (state *dummyState) setSeen(messageID imap.MessageID, seen bool) {
	state.lock.Lock()
	defer state.lock.Unlock()

	state.messages[messageID].seen = seen
}

func (state *dummyState) setFlagged(messageID imap.MessageID, flagged bool) {
	state.lock.Lock()
	defer state.lock.Unlock()

	state.messages[messageID].flagged = flagged
}

func (state *dummyState) setForwarded(messageID imap.MessageID, forwarded bool) {
	state.lock.Lock()
	defer state.lock.Unlock()

	state.messages[messageID].forwarded = forwarded
}

func (state *dummyState) isSeen(messageID imap.MessageID) bool {
	state.lock.Lock()
	defer state.lock.Unlock()

	return state.messages[messageID].seen
}

func (state *dummyState) isFlagged(messageID imap.MessageID) bool {
	state.lock.Lock()
	defer state.lock.Unlock()

	return state.messages[messageID].flagged
}

func (state *dummyState) isDraft(messageID imap.MessageID) bool {
	state.lock.Lock()
	defer state.lock.Unlock()

	// NOTE: If we want to support a custom draft folder, handle it here and update `getMessageFlags`.
	return false
}

func (state *dummyState) getMessageFlags(messageID imap.MessageID) imap.FlagSet {
	state.lock.Lock()
	defer state.lock.Unlock()

	msg := state.messages[messageID]

	flags := imap.NewFlagSet()

	if msg.seen {
		flags.AddToSelf(imap.FlagSeen)
	}

	if msg.flagged {
		flags.AddToSelf(imap.FlagFlagged)
	}

	if msg.forwarded {
		flags.AddToSelf(imap.XFlagForwarded)
		flags.AddToSelf(imap.XFlagDollarForwarded)
	}

	return flags
}

func (state *dummyState) toMailbox(mboxID imap.MailboxID) imap.Mailbox {
	return imap.Mailbox{
		ID:             mboxID,
		Name:           state.mailboxes[mboxID].mboxName,
		Flags:          state.flags,
		PermanentFlags: state.permFlags,
		Attributes:     state.attrs,
	}
}

func (state *dummyState) toMessage(messageID imap.MessageID) imap.Message {
	flags := imap.NewFlagSet()

	if state.messages[messageID].seen {
		flags.AddToSelf(imap.FlagSeen)
	}

	if state.messages[messageID].flagged {
		flags.AddToSelf(imap.FlagFlagged)
	}

	flags.AddFlagSetToSelf(state.messages[messageID].flags)

	return imap.Message{
		ID:    messageID,
		Flags: flags,
		Date:  state.messages[messageID].date,
	}
}
