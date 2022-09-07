package state

import (
	"context"
	"fmt"
	"sync"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db"
	"github.com/ProtonMail/gluon/internal/db/ent"
	"github.com/ProtonMail/gluon/internal/ids"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/bradenaw/juniper/xslices"
)

type responderStateUpdate struct {
	SnapFilter
	responders []Responder
}

func (r *responderStateUpdate) Apply(ctx context.Context, tx *ent.Tx, s *State) error {
	return s.PushResponder(ctx, tx, r.responders...)
}

func (r *responderStateUpdate) String() string {
	return fmt.Sprintf("ResponderStateUpdate: %v Responders=%v",
		r.SnapFilter,
		xslices.Map(r.responders, func(rsp Responder) string {
			return rsp.String()
		}),
	)
}

// responderDBUpdate may be returned by a responder in order to avoid locking the database writer for each update.
// Seeing as this is only used right now to clear the recent flags, we can avoid a lot of necessary database locking
// and transaction overhead.
type responderDBUpdate interface {
	apply(ctx context.Context, tx *ent.Tx) error
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
	handle(snap *snapshot, stateID StateID) ([]response.Response, responderDBUpdate, error)

	// getMessageID returns the message ID that this Responder targets.
	getMessageID() imap.InternalMessageID

	String() string
}

type exists struct {
	messageID  ids.MessageIDPair
	messageUID imap.UID
	flags      imap.FlagSet
}

func newExists(messageID ids.MessageIDPair, messageUID imap.UID, flags imap.FlagSet) *exists {
	return &exists{messageID: messageID, messageUID: messageUID, flags: flags}
}

func (u *exists) String() string {
	return fmt.Sprintf("Exists: message=%v remote=%v", u.messageID.InternalID.ShortID(), u.messageID.RemoteID)
}

type clearRecentFlagRespUpdate struct {
	messageID imap.InternalMessageID
	mboxID    imap.InternalMailboxID
}

func (u *clearRecentFlagRespUpdate) apply(ctx context.Context, tx *ent.Tx) error {
	return db.ClearRecentFlag(ctx, tx, u.mboxID, u.messageID)
}

// targetedExists needs to be separate so that we update the targetStateID safely when doing concurrent updates
// in different states. This way we also avoid the extra step of copying the `exists` data.
type targetedExists struct {
	resp          *exists
	targetStateID StateID
}

func (u *targetedExists) handle(snap *snapshot, stateID StateID) ([]response.Response, responderDBUpdate, error) {
	if snap.hasMessage(u.resp.messageID.InternalID) {
		return nil, nil, nil
	}

	var flags imap.FlagSet
	if u.targetStateID != stateID {
		flags = flags.Remove(imap.FlagRecent)
	} else {
		flags = u.resp.flags
	}

	if err := snap.appendMessage(u.resp.messageID, u.resp.messageUID, flags); err != nil {
		return nil, nil, err
	}

	seq, err := snap.getMessageSeq(u.resp.messageID.InternalID)
	if err != nil {
		return nil, nil, err
	}

	res := []response.Response{response.Exists().WithCount(seq)}

	var dbUpdate responderDBUpdate

	if recent := len(snap.getMessagesWithFlag(imap.FlagRecent)); recent > 0 {
		if flags.Contains(imap.FlagRecent) {
			dbUpdate = &clearRecentFlagRespUpdate{
				mboxID:    snap.mboxID.InternalID,
				messageID: u.resp.messageID.InternalID,
			}
		}

		res = append(res, response.Recent().WithCount(recent))
	}

	return res, dbUpdate, nil
}

func (u *targetedExists) getMessageID() imap.InternalMessageID {
	return u.resp.messageID.InternalID
}

func (u *targetedExists) String() string {
	return fmt.Sprintf("TargetedExists: message = %v remote = %v targetStateID = %v", u.resp.messageID.InternalID.ShortID(), u.resp.messageID.RemoteID, u.targetStateID)
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
}

func NewExistsStateUpdate(mailboxID imap.InternalMailboxID, messageIDs []ids.MessageIDPair, uids map[imap.InternalMessageID]*ent.UID, s *State) Update {
	var stateID StateID

	var targetStateSet bool

	if s != nil {
		stateID = s.StateID
		targetStateSet = true
	}

	responders := xslices.Map(messageIDs, func(messageID ids.MessageIDPair) *exists {
		uid := uids[messageID.InternalID]
		exists := newExists(ids.NewMessageIDPair(uid.Edges.Message), uid.UID, db.NewFlagSet(uid, uid.Edges.Message.Edges.Flags))

		return exists
	})

	return &ExistsStateUpdate{
		MBoxIDStateFilter: MBoxIDStateFilter{MboxID: mailboxID},
		responders:        responders,
		targetStateID:     stateID,
		targetStateSet:    targetStateSet,
	}
}

func newExistsStateUpdateWithExists(mailboxID imap.InternalMailboxID, responders []*exists, s *State) Update {
	var stateID StateID

	var targetStateSet bool

	if s != nil {
		stateID = s.StateID
		targetStateSet = true
	}

	return &ExistsStateUpdate{
		MBoxIDStateFilter: MBoxIDStateFilter{MboxID: mailboxID},
		responders:        responders,
		targetStateID:     stateID,
		targetStateSet:    targetStateSet,
	}
}

func (e *ExistsStateUpdate) Apply(ctx context.Context, tx *ent.Tx, s *State) error {
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

	return s.PushResponder(ctx, tx, xslices.Map(e.responders, func(e *exists) Responder {
		return &targetedExists{
			resp:          e,
			targetStateID: targetStateID,
		}
	})...)
}

func (e *ExistsStateUpdate) String() string {
	return fmt.Sprintf("ExistsStateUpdate: %v Responders = %v targetStateId = %v",
		e.MBoxIDStateFilter,
		xslices.Map(e.responders, func(rsp *exists) string {
			return rsp.String()
		}),
		e.targetStateID,
	)
}

type expunge struct {
	messageID imap.InternalMessageID
	asClose   bool
}

func NewExpunge(messageID imap.InternalMessageID, asClose bool) *expunge {
	return &expunge{
		messageID: messageID,
		asClose:   asClose,
	}
}

func (u *expunge) handle(snap *snapshot, _ StateID) ([]response.Response, responderDBUpdate, error) {
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
	if u.asClose {
		return nil, nil, nil
	}

	return []response.Response{response.Expunge(seq)}, nil, nil
}

func (u *expunge) getMessageID() imap.InternalMessageID {
	return u.messageID
}

func (u *expunge) String() string {
	return fmt.Sprintf("Expung: message = %v closed = %v",
		u.messageID.ShortID(),
		u.asClose,
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

func (u *fetch) handle(snap *snapshot, _ StateID) ([]response.Response, responderDBUpdate, error) {
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
		newMessageFlags = u.flags
	}

	if u.cameFromDifferentMailbox {
		newMessageFlags = newMessageFlags.Set(imap.FlagDeleted, curFlags.Contains(imap.FlagDeleted))
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
