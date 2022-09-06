package state

import (
	"context"
	"fmt"

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

func (r responderStateUpdate) Apply(ctx context.Context, tx *ent.Tx, s *State) error {
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
	handle(ctx context.Context, tx *ent.Tx, snap *snapshot) ([]response.Response, error)

	// getMessageID returns the message ID that this Responder targets.
	getMessageID() imap.InternalMessageID

	String() string
}

type exists struct {
	messageID  imap.InternalMessageID
	messageUID int
}

func NewExists(messageID imap.InternalMessageID, messageUID int) *exists {
	return &exists{messageID: messageID, messageUID: messageUID}
}

func (u *exists) handle(ctx context.Context, tx *ent.Tx, snap *snapshot) ([]response.Response, error) {
	if snap.hasMessage(u.messageID) {
		return nil, nil
	}

	client := tx.Client()

	remoteID, err := db.GetRemoteMessageID(ctx, client, u.messageID)
	if err != nil {
		return nil, err
	}

	if err := snap.appendMessage(ctx, client, ids.MessageIDPair{InternalID: u.messageID, RemoteID: remoteID}); err != nil {
		return nil, err
	}

	seq, err := snap.getMessageSeq(u.messageID)
	if err != nil {
		return nil, err
	}

	res := []response.Response{response.Exists().WithCount(seq)}

	if recent := len(snap.getMessagesWithFlag(imap.FlagRecent)); recent > 0 {
		if msgFlags, err := snap.getMessageFlags(u.messageID); err != nil {
			return nil, err
		} else if msgFlags.Contains(imap.FlagRecent) {
			if err := db.ClearRecentFlag(ctx, tx, snap.mboxID.InternalID, u.messageID); err != nil {
				return nil, err
			}
		}

		res = append(res, response.Recent().WithCount(recent))
	}

	return res, nil
}

func (u *exists) getMessageID() imap.InternalMessageID {
	return u.messageID
}

func (u *exists) String() string {
	return fmt.Sprintf("Exists: message=%v", u.messageID.ShortID())
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

func (u *expunge) handle(ctx context.Context, tx *ent.Tx, snap *snapshot) ([]response.Response, error) {
	if !snap.hasMessage(u.messageID) {
		return nil, nil
	}

	seq, err := snap.getMessageSeq(u.messageID)
	if err != nil {
		return nil, err
	}

	if err := snap.expungeMessage(u.messageID); err != nil {
		return nil, err
	}

	// When handling a CLOSE command, EXPUNGE responses are not sent.
	if u.asClose {
		return nil, nil
	}

	return []response.Response{response.Expunge(seq)}, nil
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

func (u *fetch) handle(ctx context.Context, tx *ent.Tx, snap *snapshot) ([]response.Response, error) {
	if !snap.hasMessage(u.messageID) {
		return nil, nil
	}

	// Get the flags in this particular snapshot (might contain Recent flag).
	curFlags, err := snap.getMessageFlags(u.messageID)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	// Get the updated newFlags in this particular snapshot (might contain Recent flag).
	newFlags, err := snap.getMessageFlags(u.messageID)
	if err != nil {
		return nil, err
	}

	// If the flags are unchanged, we don't send a FETCH response.
	if curFlags.Equals(newFlags) {
		return nil, nil
	}

	// When handling a SILENT STORE command, the FETCH response is not sent.
	if u.asSilent {
		return nil, nil
	}

	items := []response.Item{response.ItemFlags(newFlags)}

	// When handling any UID command, we should always include the message's UID.
	if u.asUID {
		uid, err := snap.getMessageUID(u.messageID)
		if err != nil {
			return nil, err
		}

		items = append(items, response.ItemUID(uid))
	}

	seq, err := snap.getMessageSeq(u.messageID)
	if err != nil {
		return nil, err
	}

	return []response.Response{response.Fetch(seq).WithItems(items...)}, nil
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
