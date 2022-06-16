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

	messages   map[string]*dummyMessage
	labels     map[string]*dummyLabel
	lastIMAPID imap.ID

	lock sync.RWMutex
}

type dummyLabel struct {
	labelName []string
	exclusive bool
}

type dummyMessage struct {
	literal []byte
	seen    bool
	flagged bool
	date    time.Time

	labelIDs map[string]struct{}
}

func newDummyState(flags, permFlags, attrs imap.FlagSet) *dummyState {
	return &dummyState{
		flags:      flags,
		permFlags:  permFlags,
		attrs:      attrs,
		messages:   make(map[string]*dummyMessage),
		labels:     make(map[string]*dummyLabel),
		lastIMAPID: imap.NewID(),
	}
}

func (state *dummyState) recordIMAPID(ctx context.Context) {
	if id, ok := imap.GetIMAPIDFromContext(ctx); ok {
		state.lock.Lock()
		defer state.lock.Unlock()
		state.lastIMAPID = id
	}
}

func (state *dummyState) getLabels() []imap.Mailbox {
	state.lock.Lock()
	defer state.lock.Unlock()

	return xslices.Map(maps.Keys(state.labels), func(labelID string) imap.Mailbox {
		return state.toMailbox(labelID)
	})
}

func (state *dummyState) getLabel(labelID string) imap.Mailbox {
	state.lock.Lock()
	defer state.lock.Unlock()

	return state.toMailbox(labelID)
}

func (state *dummyState) createLabel(name []string, exclusive bool) imap.Mailbox {
	state.lock.Lock()
	defer state.lock.Unlock()

	labelID := uuid.NewString()

	state.labels[labelID] = &dummyLabel{
		labelName: name,
		exclusive: exclusive,
	}

	return state.toMailbox(labelID)
}

func (state *dummyState) updateLabel(labelID string, name []string) {
	state.lock.Lock()
	defer state.lock.Unlock()

	state.labels[labelID].labelName = name
}

func (state *dummyState) deleteLabel(labelID string) {
	state.lock.Lock()
	defer state.lock.Unlock()

	delete(state.labels, labelID)

	for _, message := range state.messages {
		delete(message.labelIDs, labelID)
	}
}

func (state *dummyState) getMessages() []imap.Message {
	state.lock.Lock()
	defer state.lock.Unlock()

	return xslices.Map(maps.Keys(state.messages), func(messageID string) imap.Message {
		return state.toMessage(messageID)
	})
}

func (state *dummyState) getMessage(messageID string) imap.Message {
	state.lock.Lock()
	defer state.lock.Unlock()

	return state.toMessage(messageID)
}

func (state *dummyState) getLabelIDs(messageID string) []string {
	state.lock.Lock()
	defer state.lock.Unlock()

	return maps.Keys(state.messages[messageID].labelIDs)
}

func (state *dummyState) getLiteral(messageID string) []byte {
	state.lock.Lock()
	defer state.lock.Unlock()

	return state.messages[messageID].literal
}

func (state *dummyState) createMessage(mboxID string, literal []byte, seen, flagged bool, date time.Time) imap.Message {
	state.lock.Lock()
	defer state.lock.Unlock()

	messageID := uuid.NewString()

	state.messages[messageID] = &dummyMessage{
		literal:  literal,
		seen:     seen,
		flagged:  flagged,
		date:     date,
		labelIDs: map[string]struct{}{mboxID: {}},
	}

	return state.toMessage(messageID)
}

func (state *dummyState) labelMessage(messageID, labelID string) {
	state.lock.Lock()
	defer state.lock.Unlock()

	if state.labels[labelID].exclusive {
		state.messages[messageID].labelIDs = make(map[string]struct{})
	}

	state.messages[messageID].labelIDs[labelID] = struct{}{}
}

func (state *dummyState) unlabelMessage(messageID, labelID string) {
	state.lock.Lock()
	defer state.lock.Unlock()

	delete(state.messages[messageID].labelIDs, labelID)
}

func (state *dummyState) setSeen(messageID string, seen bool) {
	state.lock.Lock()
	defer state.lock.Unlock()

	state.messages[messageID].seen = seen
}

func (state *dummyState) setFlagged(messageID string, flagged bool) {
	state.lock.Lock()
	defer state.lock.Unlock()

	state.messages[messageID].flagged = flagged
}

func (state *dummyState) isSeen(messageID string) bool {
	state.lock.Lock()
	defer state.lock.Unlock()

	return state.messages[messageID].seen
}

func (state *dummyState) isFlagged(messageID string) bool {
	state.lock.Lock()
	defer state.lock.Unlock()

	return state.messages[messageID].flagged
}

func (state *dummyState) toMailbox(labelID string) imap.Mailbox {
	return imap.Mailbox{
		ID:             labelID,
		Name:           state.labels[labelID].labelName,
		Flags:          state.flags,
		PermanentFlags: state.permFlags,
		Attributes:     state.attrs,
	}
}

func (state *dummyState) toMessage(messageID string) imap.Message {
	flags := imap.NewFlagSet()

	if state.messages[messageID].seen {
		flags = flags.Add(imap.FlagSeen)
	}

	if state.messages[messageID].flagged {
		flags = flags.Add(imap.FlagFlagged)
	}

	return imap.Message{
		ID:    messageID,
		Flags: flags,
		Date:  state.messages[messageID].date,
	}
}
