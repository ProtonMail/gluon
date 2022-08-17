package backend

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/backend/ent"
	"github.com/ProtonMail/gluon/internal/remote"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/bradenaw/juniper/sets"
	"github.com/bradenaw/juniper/xslices"
)

// TODO(REFACTOR): Decide on the best way to pass around user/state/snap objects! Currently quite gross...
type State struct {
	*user

	stateID    int
	metadataID remote.ConnMetadataID

	idleCh   chan response.Response
	idleLock sync.RWMutex

	res     []responder
	resLock sync.Mutex

	snap *snapshot
	ro   bool

	doneCh chan struct{}
	stopCh chan struct{}
}

func (state *State) UserID() string {
	return state.userID
}

func (state *State) List(ctx context.Context, ref, pattern string, subscribed bool, fn func(map[string]Match) error) error {
	return state.tx(ctx, func(tx *ent.Tx) error {
		mailboxes, err := DBGetAllMailboxes(ctx, tx.Client())
		if err != nil {
			return err
		}

		matches, err := getMatches(ctx, mailboxes, ref, pattern, state.delimiter, subscribed)
		if err != nil {
			return err
		}

		return fn(matches)
	})
}

func (state *State) Select(ctx context.Context, name string, fn func(*Mailbox) error) error {
	return state.tx(ctx, func(tx *ent.Tx) error {
		mbox, err := DBGetMailboxByName(ctx, tx.Client(), name)
		if err != nil {
			return ErrNoSuchMailbox
		}

		if state.snap != nil {
			if err := state.close(ctx, tx); err != nil {
				return err
			}
		}

		snap, err := newSnapshot(ctx, state, mbox)
		if err != nil {
			return err
		}

		if err := DBClearRecentFlags(ctx, tx, mbox.MailboxID); err != nil {
			return err
		}

		state.snap = snap
		state.ro = false

		return fn(newMailbox(tx, mbox, state, state.snap))
	})
}

func (state *State) Examine(ctx context.Context, name string, fn func(*Mailbox) error) error {
	return state.tx(ctx, func(tx *ent.Tx) error {
		mbox, err := DBGetMailboxByName(ctx, tx.Client(), name)
		if err != nil {
			return ErrNoSuchMailbox
		}

		if state.snap != nil {
			if err := state.close(ctx, tx); err != nil {
				return err
			}
		}

		snap, err := newSnapshot(ctx, state, mbox)
		if err != nil {
			return err
		}

		state.snap = snap
		state.ro = true

		return fn(newMailbox(tx, mbox, state, state.snap))
	})
}

func (state *State) Create(ctx context.Context, name string) error {
	return state.tx(ctx, func(tx *ent.Tx) error {
		// If the mailbox name is suffixed with the server's hierarchy separator, remove the separator and still create
		// the mailbox
		if strings.HasSuffix(name, state.delimiter) {
			name = strings.TrimRight(name, state.delimiter)
		}

		if exists, err := DBMailboxExistsWithName(ctx, tx.Client(), name); err != nil {
			return err
		} else if exists {
			return ErrExistingMailbox
		}

		for _, superior := range listSuperiors(name, state.delimiter) {
			if exists, err := DBMailboxExistsWithName(ctx, tx.Client(), superior); err != nil {
				return err
			} else if exists {
				continue
			}

			if err := state.actionCreateMailbox(ctx, tx, superior); err != nil {
				return err
			}
		}

		if err := state.actionCreateMailbox(ctx, tx, name); err != nil {
			return err
		}

		return nil
	})
}

func (state *State) Delete(ctx context.Context, name string) error {
	return state.tx(ctx, func(tx *ent.Tx) error {
		mbox, err := DBGetMailboxByName(ctx, tx.Client(), name)
		if err != nil {
			return ErrNoSuchMailbox
		}

		return state.actionDeleteMailbox(ctx, tx, mbox.RemoteID, name)
	})
}

func (state *State) Rename(ctx context.Context, oldName, newName string) error {
	return state.tx(ctx, func(tx *ent.Tx) error {
		client := tx.Client()
		mbox, err := DBGetMailboxByName(ctx, client, oldName)
		if err != nil {
			return ErrNoSuchMailbox
		}

		if exists, err := DBMailboxExistsWithName(ctx, client, newName); err != nil {
			return err
		} else if exists {
			return ErrExistingMailbox
		}

		for _, superior := range listSuperiors(newName, state.delimiter) {
			if exists, err := DBMailboxExistsWithName(ctx, client, superior); err != nil {
				return err
			} else if exists {
				if superior == oldName {
					return ErrExistingMailbox
				}
				continue
			}

			if err := state.actionCreateMailbox(ctx, tx, superior); err != nil {
				return err
			}
		}

		if oldName == imap.Inbox {
			return state.renameInbox(ctx, tx, mbox, newName)
		}

		mailboxes, err := DBGetAllMailboxes(ctx, client)
		if err != nil {
			return err
		}

		inferiors := listInferiors(oldName, state.delimiter, xslices.Map(mailboxes, func(mailbox *ent.Mailbox) string {
			return mailbox.Name
		}))

		for _, inferior := range inferiors {
			mbox, err := DBGetMailboxByName(ctx, client, inferior)
			if err != nil {
				return ErrNoSuchMailbox
			}

			newInferior := newName + strings.TrimPrefix(inferior, oldName)

			if err := state.actionUpdateMailbox(ctx, tx, mbox.RemoteID, inferior, newInferior); err != nil {
				return err
			}
		}

		return state.actionUpdateMailbox(ctx, tx, mbox.RemoteID, oldName, newName)
	})
}

func (state *State) Subscribe(ctx context.Context, name string) error {
	return state.tx(ctx, func(tx *ent.Tx) error {
		mbox, err := DBGetMailboxByName(ctx, tx.Client(), name)
		if err != nil {
			return ErrNoSuchMailbox
		} else if mbox.Subscribed {
			return ErrAlreadySubscribed
		}

		return mbox.Update().SetSubscribed(true).Exec(ctx)
	})
}

func (state *State) Unsubscribe(ctx context.Context, name string) error {
	return state.tx(ctx, func(tx *ent.Tx) error {
		mbox, err := DBGetMailboxByName(ctx, tx.Client(), name)
		if err != nil {
			return ErrNoSuchMailbox
		} else if !mbox.Subscribed {
			return ErrAlreadyUnsubscribed
		}

		return mbox.Update().SetSubscribed(false).Exec(ctx)
	})
}

func (state *State) Idle(ctx context.Context, fn func([]response.Response, chan response.Response) error) error {
	res, err := state.beginIdle(ctx)
	if err != nil {
		return err
	}

	defer state.endIdle()

	return fn(res, state.idleCh)
}

func (state *State) Mailbox(ctx context.Context, name string, fn func(*Mailbox) error) error {
	return state.tx(ctx, func(tx *ent.Tx) error {
		mbox, err := DBGetMailboxByName(ctx, tx.Client(), name)
		if err != nil {
			return ErrNoSuchMailbox
		}

		if state.snap != nil && state.snap.mboxID.InternalID == mbox.MailboxID {
			return fn(newMailbox(tx, mbox, state, state.snap))
		}

		snap, err := newSnapshot(ctx, state, mbox)
		if err != nil {
			return err
		}

		return fn(newMailbox(tx, mbox, state, snap))
	})
}

func (state *State) Selected(ctx context.Context, fn func(*Mailbox) error) error {
	if !state.IsSelected() {
		return ErrSessionNotSelected
	}

	return state.tx(ctx, func(tx *ent.Tx) error {
		mbox, err := DBGetMailboxByID(ctx, tx.Client(), state.snap.mboxID.InternalID)
		if err != nil {
			return err
		}

		return fn(newMailbox(tx, mbox, state, state.snap))
	})
}

func (state *State) IsSelected() bool {
	return state.snap != nil
}

func (state *State) SetConnMetadataKeyValue(key string, value any) error {
	return state.remote.SetConnMetadataValue(state.metadataID, key, value)
}

func (state *State) Done() <-chan struct{} {
	if state == nil {
		return nil
	}

	return state.doneCh
}

func (state *State) Close(ctx context.Context) error {
	defer close(state.stopCh)

	return state.removeState(ctx, state.stateID)
}

// renameInbox creates a new mailbox and moves everything there.
func (state *State) renameInbox(ctx context.Context, tx *ent.Tx, inbox *ent.Mailbox, newName string) error {
	mbox, err := state.actionCreateAndGetMailbox(ctx, tx, newName)
	if err != nil {
		return err
	}

	messages, err := DBGetMailboxMessages(ctx, inbox)
	if err != nil {
		return err
	}

	messageIDs := xslices.Map(messages, func(messageUID *ent.UID) MessageIDPair {
		return NewMessageIDPair(messageUID.Edges.Message)
	})

	mboxIDPair := NewMailboxIDPair(mbox)

	if _, err := state.actionAddMessagesToMailbox(ctx, tx, messageIDs, mboxIDPair); err != nil {
		return err
	}

	if err := state.actionRemoveMessagesFromMailbox(ctx, tx, messageIDs, NewMailboxIDPair(inbox)); err != nil {
		return err
	}

	return nil
}

func (state *State) beginIdle(ctx context.Context) ([]response.Response, error) {
	var res []response.Response

	if err := state.tx(ctx, func(tx *ent.Tx) error {
		state.idleLock.Lock()
		defer state.idleLock.Unlock()

		var err error

		if res, err = state.flushResponses(ctx, tx, true); err != nil {
			return err
		}

		state.idleCh = make(chan response.Response)

		return nil
	}); err != nil {
		return nil, err
	}

	return res, nil
}

func (state *State) endIdle() {
	state.idleLock.Lock()
	defer state.idleLock.Unlock()

	close(state.idleCh)

	state.idleCh = nil
}

func (state *State) getLiteral(messageID imap.InternalMessageID) ([]byte, error) {
	return state.store.Get(messageID)
}

func (state *State) flushResponses(ctx context.Context, tx *ent.Tx, permitExpunge bool) ([]response.Response, error) {
	var responses []response.Response

	for _, responder := range state.popResponders(permitExpunge) {
		res, err := responder.handle(ctx, tx, state.snap)
		if err != nil {
			return nil, err
		}

		responses = append(responses, res...)
	}

	return responses, nil
}

func (state *State) pushResponder(ctx context.Context, tx *ent.Tx, responder ...responder) error {
	state.idleLock.RLock()
	defer state.idleLock.RUnlock()

	if state.idleCh == nil {
		return state.queueResponder(ctx, tx, responder...)
	}

	for _, responder := range responder {
		res, err := responder.handle(ctx, tx, state.snap)
		if err != nil {
			return err
		}

		for _, res := range res {
			state.idleCh <- res
		}
	}

	return nil
}

func (state *State) queueResponder(ctx context.Context, tx *ent.Tx, responder ...responder) error {
	state.resLock.Lock()
	defer state.resLock.Unlock()

	state.res = append(state.res, responder...)

	return nil
}

// popResponders pops responders from the state.
// If permitExpunge is false, it stops when it encounters the first expunge responder.
// This is important: suppose someone first expunges a message (generating an EXPUNGE)
// then puts it back (generating an EXISTS). If we didn't stop at the first expunge,
// we would send the EXISTS to the client, followed by an expunge afterwards (wrong).
func (state *State) popResponders(permitExpunge bool) []responder {
	state.resLock.Lock()
	defer state.resLock.Unlock()

	if len(state.res) == 0 {
		return nil
	}

	var pop, rem []responder

	skipIDs := make(sets.Map[imap.InternalMessageID])

	for _, res := range state.res {
		if permitExpunge {
			pop = append(pop, res)
		} else if _, ok := res.(*expunge); ok {
			rem = append(rem, res)
			skipIDs.Add(res.getMessageID())
		} else if _, ok := res.(*exists); ok {
			if skipIDs.Contains(res.getMessageID()) {
				rem = append(rem, res)
				skipIDs.Remove(res.getMessageID())
			} else {
				pop = append(pop, res)
			}
		} else {
			pop = append(pop, res)
		}
	}

	state.res = rem

	return pop
}

func (state *State) updateMailboxRemoteID(internalID imap.InternalMailboxID, remoteID imap.LabelID) error {
	if state.snap != nil {
		if err := state.snap.updateMailboxRemoteID(internalID, remoteID); err != nil {
			return err
		}
	}

	return nil
}

func (state *State) updateMessageRemoteID(internalID imap.InternalMessageID, remoteID imap.MessageID) error {
	if state.snap != nil && state.snap.hasMessage(internalID) {
		if err := state.snap.updateMessageRemoteID(internalID, remoteID); err != nil {
			return err
		}
	}

	return nil
}

// We don't want the queue closed to be reported as an error.
// User will clean up existing metadata entries by itself when closed.
func (state *State) deleteConnMetadata() error {
	if err := state.remote.DeleteConnMetadataStore(state.metadataID); err != nil {
		return fmt.Errorf("failed to delete conn metadata store: %w", err)
	}

	return nil
}

func (state *State) close(ctx context.Context, tx *ent.Tx) error {
	state.snap = nil

	state.res = nil

	return nil
}
