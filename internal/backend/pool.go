package backend

import (
	"context"
	"sync"

	"github.com/ProtonMail/gluon/internal/backend/ent"
	"github.com/bradenaw/juniper/xslices"
	"golang.org/x/exp/slices"
)

// pool keeps track of snapshots that are preventing a message from being fully removed from the database.
type pool struct {
	entries map[string]map[string][]*snapshot
	lock    sync.RWMutex
}

func newPool() *pool {
	return &pool{
		entries: make(map[string]map[string][]*snapshot),
	}
}

// hasSnap returns whether there are messages that cannot be removed yet because
// the given snapshot has still to be notified.
func (pool *pool) hasSnap(snap *snapshot) bool {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	mbox, ok := pool.entries[snap.mboxID]
	if !ok {
		return false
	}

	for _, message := range mbox {
		if slices.Contains(message, snap) {
			return true
		}
	}

	return false
}

// hasMessage returns whether there are snapshots that have still to be notified
// about the removal of the given message in the given mailbox.
func (pool *pool) hasMessage(mboxID, messageID string) bool {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	mbox, ok := pool.entries[mboxID]
	if !ok {
		return false
	}

	return len(mbox[messageID]) > 0
}

// addMessage adds the given message and snapshot to the deletion pool.
// This means that the given snapshot has not yet been notified of the message's removal.
func (pool *pool) addMessage(snap *snapshot, messageID string) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	mbox, ok := pool.entries[snap.mboxID]
	if !ok {
		pool.entries[snap.mboxID] = make(map[string][]*snapshot)
		mbox = pool.entries[snap.mboxID]
	}

	if !slices.Contains(mbox[messageID], snap) {
		mbox[messageID] = append(mbox[messageID], snap)
	}
}

// removeMessages removes the given messages for the given mailbox from the deletion pool.
func (pool *pool) removeMessages(mboxID string, messageIDs ...string) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	mbox, ok := pool.entries[mboxID]
	if !ok {
		return ErrNoSuchMailbox
	}

	for _, messageID := range messageIDs {
		if delete(mbox, messageID); len(mbox) == 0 {
			delete(pool.entries, mboxID)
		}
	}

	return nil
}

// expungeMessage indicates that the given snapshot has been notified of the message's removal.
// If all snapshots have been notified, the message can finally be removed from the mailbox.
func (pool *pool) expungeMessage(ctx context.Context, tx *ent.Tx, snap *snapshot, messageID string) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	mbox, ok := pool.entries[snap.mboxID]
	if !ok {
		return ErrNoSuchMailbox
	}

	snaps, ok := mbox[messageID]
	if !ok {
		return ErrNoSuchMessage
	}

	idx := slices.Index(snaps, snap)
	if idx < 0 {
		return ErrNoSuchSnapshot
	}

	if snaps = xslices.Remove(snaps, idx, 1); len(snaps) == 0 {
		if err := txRemoveMessagesFromMailbox(ctx, tx, []string{messageID}, snap.mboxID); err != nil {
			return err
		}

		if delete(mbox, messageID); len(mbox) == 0 {
			delete(pool.entries, snap.mboxID)
		}
	} else {
		mbox[messageID] = snaps
	}

	return nil
}

// expungeSnap indicates that the given snapshot has been closed.
// Messages that were waiting for its notification no longer have to wait.
func (pool *pool) expungeSnap(ctx context.Context, tx *ent.Tx, snap *snapshot) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	mbox, ok := pool.entries[snap.mboxID]
	if !ok {
		return ErrNoSuchMailbox
	}

	var messageIDs []string

	for messageID, snaps := range mbox {
		idx := slices.Index(snaps, snap)
		if idx < 0 {
			continue
		}

		if snaps = xslices.Remove(snaps, idx, 1); len(snaps) > 0 {
			mbox[messageID] = snaps
		} else {
			messageIDs = append(messageIDs, messageID)
		}
	}

	if err := txRemoveMessagesFromMailbox(ctx, tx, messageIDs, snap.mboxID); err != nil {
		return err
	}

	for _, messageID := range messageIDs {
		if delete(mbox, messageID); len(mbox) == 0 {
			delete(pool.entries, snap.mboxID)
		}
	}

	return nil
}

func (pool *pool) updateMailboxID(oldID, newID string) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	if mbox, ok := pool.entries[oldID]; ok {
		pool.entries[newID] = mbox
		delete(pool.entries, oldID)
	}
}

func (pool *pool) updateMessageID(oldID, newID string) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	for _, mbox := range pool.entries {
		if snaps, ok := mbox[oldID]; ok {
			mbox[newID] = snaps
			delete(mbox, oldID)
		}
	}
}
