package state

import (
	"context"
	"fmt"
	"github.com/ProtonMail/gluon/db"
	"sync"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/contexts"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/bradenaw/juniper/xslices"
)

type responderStateUpdate struct {
	SnapFilter
	responders []Responder
}

func (r *responderStateUpdate) Apply(ctx context.Context, tx db.Transaction, s *State) error {
	return s.PushResponder(ctx, tx, r.responders...)
}

func (r *responderStateUpdate) String() string {
	return fmt.Sprintf("ResponderStateUpdate: %v Responders=%v",
		r.SnapFilter.String(),
		xslices.Map(r.responders, func(rsp Responder) string {
			return rsp.String()
		}),
	)
}

// responderDBUpdate may be returned by a responder in order to avoid locking the database writer for each update.
// Seeing as this is only used right now to clear the recent flags, we can avoid a lot of necessary database locking
// and transaction overhead.
type responderDBUpdate interface {
	apply(ctx context.Context, tx db.Transaction) error
}

func NewMailboxIDResponderStateUpdate(id imap.InternalMailboxID, responders ...Responder) Update {
	return &responderStateUpdate{SnapFilter: NewMBoxIDStateFilter(id), responders: responders}
}

func NewMessageIDResponderStateUpdate(id imap.InternalMessageID, responders ...Responder) Update {
	return &responderStateUpdate{SnapFilter: NewMessageIDStateFilter(id), responders: responders}
}

func NewMessageIDAndMailboxIDResponderStateUpdate(messageID imap.InternalMessageID, mboxID imap.InternalMailboxID, responders ...Responder) Update {
	return &responderStateUpdate{SnapFilter: NewMessageAndMBoxIDStateFilter(messageID, mboxID), responders: responders}
}

type Responder interface {
	// handle generates responses in the context of the given snapshot.
	handle(ctx context.Context, snap *snapshot, stateID StateID) ([]response.Response, responderDBUpdate, error)

	// getMessageID returns the message ID that this Responder targets.
	getMessageID() imap.InternalMessageID

	String() string
}

type exists struct {
	messageID  db.MessageIDPair
	messageUID imap.UID
	flags      imap.FlagSet
}

func newExists(messageID db.MessageIDPair, messageUID imap.UID, flags imap.FlagSet) *exists {
	return &exists{messageID: messageID, messageUID: messageUID, flags: flags}
}

func (u *exists) String() string {
	return fmt.Sprintf("Exists: message=%v remote=%v", u.messageID.InternalID.ShortID(), u.messageID.RemoteID)
}

type clearRecentFlagRespUpdate struct {
	messageID imap.InternalMessageID
	mboxID    imap.InternalMailboxID
}

func (u *clearRecentFlagRespUpdate) apply(ctx context.Context, tx db.Transaction) error {
	return tx.ClearRecentFlagInMailboxOnMessage(ctx, u.mboxID, u.messageID)
}

// targetedExists needs to be separate so that we update the targetStateID safely when doing concurrent updates
// in different states. This way we also avoid the extra step of copying the `exists` data.
type targetedExists struct {
	resp           *exists
	targetStateID  StateID
	originStateID  StateID
	originStateSet bool
}

func (u *targetedExists) handle(ctx context.Context, snap *snapshot, stateID StateID) ([]response.Response, responderDBUpdate, error) {
	if snap.hasMessage(u.resp.messageID.InternalID) {
		return nil, nil, nil
	}

	var flags imap.FlagSet
	if u.targetStateID != stateID {
		flags = u.resp.flags.Remove(imap.FlagRecent)
	} else {
		flags = u.resp.flags
	}

	if u.originStateSet && u.originStateID == stateID {
		if err := snap.appendMessage(u.resp.messageID, u.resp.messageUID, flags); err != nil {
			reporter.ExceptionWithContext(ctx, "Failed to append message to snap via targetedExists", reporter.Context{"error": err})
			return nil, nil, err
		}
	} else {
		if err := snap.appendMessageFromOtherState(u.resp.messageID, u.resp.messageUID, flags); err != nil {
			return nil, nil, err
		}
	}

	res := []response.Response{response.Exists().WithCount(imap.SeqID(snap.messages.len()))}

	var dbUpdate responderDBUpdate

	if recent := snap.getMessagesWithFlagCount(imap.FlagRecent); recent > 0 {
		if flags.ContainsUnchecked(imap.FlagRecentLowerCase) {
			dbUpdate = &clearRecentFlagRespUpdate{
				mboxID:    snap.mboxID.InternalID,
				messageID: u.resp.messageID.InternalID,
			}
		}

		res = append(res, response.Recent().WithCount(uint32(recent)))
	}

	return res, dbUpdate, nil
}

func (u *targetedExists) getMessageID() imap.InternalMessageID {
	return u.resp.messageID.InternalID
}

func (u *targetedExists) String() string {
	var originState string

	if u.originStateSet {
		originState = fmt.Sprintf("%v", u.originStateID)
	} else {
		originState = "None"
	}

	return fmt.Sprintf(
		"TargetedExists: message = %v uid = %v remote = %v targetStateID = %v originStateID = %v",
		u.resp.messageID.InternalID.ShortID(),
		u.resp.messageUID,
		u.resp.messageID.RemoteID,
		u.targetStateID,
		originState,
	)
}

// ExistsStateUpdate needs to be a separate update since it has to deal with a Recent flag propagation. If a session
// with a selected state appends a message, only that state should see the recent flag. If a message is appended to a
// non-selected mailbox or arrives from remote, the first state with the selected mailbox should get the flag.
// See ExistsStateUpdate.Apply() for more info.
type ExistsStateUpdate struct {
	lock sync.Mutex
	MBoxIDStateFilter
	responders     []*exists
	targetStateID  StateID
	targetStateSet bool
	originStateID  StateID
	originStateSet bool
}

func newExistsStateUpdateWithExists(mailboxID imap.InternalMailboxID, responders []*exists, s *State) Update {
	var (
		stateID        StateID
		originStateID  StateID
		targetStateSet bool
		originStateSet bool
	)

	if s != nil {
		stateID = s.StateID
		targetStateSet = true
		originStateID = s.StateID
		originStateSet = true
	}

	return &ExistsStateUpdate{
		MBoxIDStateFilter: MBoxIDStateFilter{MboxID: mailboxID},
		responders:        responders,
		targetStateID:     stateID,
		targetStateSet:    targetStateSet,
		originStateSet:    originStateSet,
		originStateID:     originStateID,
	}
}

func (e *ExistsStateUpdate) Apply(ctx context.Context, tx db.Transaction, s *State) error {
	// This check needs to be thread safe since we don't know when a state update
	// will be executed. Before each of these updates run we check whether at state
	// target ID has been set and update for the first state that manages to run this code
	// otherwise. To avoid race conditions on the contents of the exists update we
	// create a new responder which has the correct data and avoid race conditions all together
	targetStateID := func() StateID {
		e.lock.Lock()
		defer e.lock.Unlock()

		if !e.targetStateSet {
			e.targetStateID = s.StateID
			e.targetStateSet = true
		}

		return e.targetStateID
	}()

	return s.PushResponder(ctx, tx, xslices.Map(e.responders, func(ex *exists) Responder {
		return &targetedExists{
			resp:           ex,
			targetStateID:  targetStateID,
			originStateID:  e.originStateID,
			originStateSet: e.originStateSet,
		}
	})...)
}

func (e *ExistsStateUpdate) String() string {
	var originState string
	if e.originStateSet {
		originState = fmt.Sprintf("%v", e.originStateID)
	} else {
		originState = "None"
	}

	return fmt.Sprintf("ExistsStateUpdate: %v Responders = %v targetStateID = %v originStateID = %v",
		e.MBoxIDStateFilter.String(),
		xslices.Map(e.responders, func(rsp *exists) string {
			return rsp.String()
		}),
		e.targetStateID,
		originState,
	)
}

type expunge struct {
	messageID imap.InternalMessageID
}

func NewExpunge(messageID imap.InternalMessageID) *expunge {
	return &expunge{
		messageID: messageID,
	}
}

func (u *expunge) handle(ctx context.Context, snap *snapshot, _ StateID) ([]response.Response, responderDBUpdate, error) {
	if !snap.hasMessage(u.messageID) {
		return nil, nil, nil
	}

	seq, err := snap.getMessageSeq(u.messageID)
	if err != nil {
		return nil, nil, err
	}

	if err := snap.expungeMessage(u.messageID); err != nil {
		return nil, nil, err
	}

	// When handling a CLOSE command, EXPUNGE responses are not sent.
	if contexts.IsClose(ctx) {
		return nil, nil, nil
	}

	return []response.Response{response.Expunge(seq)}, nil, nil
}

func (u *expunge) getMessageID() imap.InternalMessageID {
	return u.messageID
}

func (u *expunge) String() string {
	return fmt.Sprintf("Expunge: message = %v",
		u.messageID.ShortID(),
	)
}

const (
	FetchFlagOpAdd = iota
	FetchFlagOpRem
	FetchFlagOpSet
)

type fetch struct {
	messageID imap.InternalMessageID
	flags     imap.FlagSet

	fetchFlagOp              int
	asUID                    bool
	asSilent                 bool
	cameFromDifferentMailbox bool
}

func NewFetch(messageID imap.InternalMessageID, flags imap.FlagSet, asUID, asSilent, cameFromDifferentMailbox bool, fetchFlagOp int) *fetch {
	return &fetch{
		messageID:                messageID,
		flags:                    flags,
		asUID:                    asUID,
		asSilent:                 asSilent,
		fetchFlagOp:              fetchFlagOp,
		cameFromDifferentMailbox: cameFromDifferentMailbox,
	}
}

func (u *fetch) handle(_ context.Context, snap *snapshot, _ StateID) ([]response.Response, responderDBUpdate, error) {
	if !snap.hasMessage(u.messageID) {
		return nil, nil, nil
	}

	// Get the flags in this particular snapshot (might contain Recent flag).
	curFlags, err := snap.getMessageFlags(u.messageID)
	if err != nil {
		return nil, nil, err
	}

	// Set the new flags as per the fetch response (recent flag is preserved).
	var newMessageFlags imap.FlagSet

	switch u.fetchFlagOp {
	case FetchFlagOpAdd:
		newMessageFlags = curFlags.AddFlagSet(u.flags)
	case FetchFlagOpRem:
		newMessageFlags = curFlags.RemoveFlagSet(u.flags)
	case FetchFlagOpSet:
		newMessageFlags = u.flags.Clone()
	}

	if u.cameFromDifferentMailbox {
		newMessageFlags.SetOnSelf(imap.FlagDeleted, curFlags.ContainsUnchecked(imap.FlagDeletedLowerCase))
	}

	if err := snap.setMessageFlags(u.messageID, newMessageFlags); err != nil {
		return nil, nil, err
	}

	// Get the updated newFlags in this particular snapshot (might contain Recent flag).
	newFlags, err := snap.getMessageFlags(u.messageID)
	if err != nil {
		return nil, nil, err
	}

	// If the flags are unchanged, we don't send a FETCH response.
	if curFlags.Equals(newFlags) {
		return nil, nil, nil
	}

	// When handling a SILENT STORE command, the FETCH response is not sent.
	if u.asSilent {
		return nil, nil, nil
	}

	items := []response.Item{response.ItemFlags(newFlags)}

	// When handling any UID command, we should always include the message's UID.
	if u.asUID {
		uid, err := snap.getMessageUID(u.messageID)
		if err != nil {
			return nil, nil, err
		}

		items = append(items, response.ItemUID(uid))
	}

	seq, err := snap.getMessageSeq(u.messageID)
	if err != nil {
		return nil, nil, err
	}

	return []response.Response{response.Fetch(seq).WithItems(items...)}, nil, nil
}

func (u *fetch) getMessageID() imap.InternalMessageID {
	return u.messageID
}

func (u *fetch) String() string {
	return fmt.Sprintf("Fetch: message = %v flags = %v uid = %v silent = %v",
		u.messageID.ShortID(),
		u.flags,
		u.asUID,
		u.asSilent,
	)
}
