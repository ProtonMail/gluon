package backend

import (
	"context"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/backend/ent"
	"github.com/ProtonMail/gluon/internal/response"
)

type responder interface {
	// handle generates responses in the context of the given snapshot.
	handle(ctx context.Context, tx *ent.Tx, snap *snapshot) ([]response.Response, error)

	// getMessageID returns the message ID that this responder targets.
	getMessageID() string

	// setMessageID sets the message ID that the responder targets.
	setMessageID(string)
}

type exists struct {
	messageID  string
	messageUID int
}

func newExists(messageID string, messageUID int) *exists {
	return &exists{messageID: messageID, messageUID: messageUID}
}

func (u *exists) handle(ctx context.Context, tx *ent.Tx, snap *snapshot) ([]response.Response, error) {
	if err := snap.appendMessage(ctx, tx, u.messageID); err != nil {
		return nil, err
	}

	seq, err := snap.getMessageSeq(u.messageID)
	if err != nil {
		return nil, err
	}

	res := []response.Response{response.Exists().WithCount(seq)}

	if recent := len(snap.getMessagesWithFlag(imap.FlagRecent)); recent > 0 {
		if err := DBClearRecentFlag(ctx, tx, snap.mboxID, u.messageID); err != nil {
			return nil, err
		}

		res = append(res, response.Recent().WithCount(recent))
	}

	return res, nil
}

func (u *exists) getMessageID() string {
	return u.messageID
}

func (u *exists) setMessageID(newID string) {
	u.messageID = newID
}

type expunge struct {
	messageID string
	asClose   bool
}

func newExpunge(messageID string, asClose bool) *expunge {
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

	if err := snap.expungeMessage(ctx, tx, u.messageID); err != nil {
		return nil, err
	}

	// When handling a CLOSE command, EXPUNGE responses are not sent.
	if u.asClose {
		return nil, nil
	}

	return []response.Response{response.Expunge(seq)}, nil
}

func (u *expunge) getMessageID() string {
	return u.messageID
}

func (u *expunge) setMessageID(newID string) {
	u.messageID = newID
}

type fetch struct {
	messageID string
	flags     imap.FlagSet

	asUID    bool
	asSilent bool
}

func newFetch(messageID string, flags imap.FlagSet, asUID, asSilent bool) *fetch {
	return &fetch{
		messageID: messageID,
		flags:     flags,
		asUID:     asUID,
		asSilent:  asSilent,
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
	if err := snap.setMessageFlags(u.messageID, u.flags); err != nil {
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

func (u *fetch) getMessageID() string {
	return u.messageID
}

func (u *fetch) setMessageID(newID string) {
	u.messageID = newID
}
