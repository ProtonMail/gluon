package backend

import (
	"context"
	"fmt"
	"strings"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/backend/ent"
	"github.com/ProtonMail/gluon/internal/queue"
	"github.com/ProtonMail/gluon/internal/remote"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/bradenaw/juniper/sets"
	"github.com/bradenaw/juniper/xslices"
	"github.com/sirupsen/logrus"
)

type State struct {
	user stateUserAccessor

	stateID    int
	metadataID remote.ConnMetadataID

	idleCh chan response.Response

	res []responder

	snap *snapshot
	ro   bool

	doneCh chan struct{}

	updatesQueue *queue.QueuedChannel[stateUpdate]

	delimiter string
}

type stateUpdate interface {
	// filter returns true when the state can be passed into apply.
	filter(s *State) bool
	// apply the update to a given state.
	apply(cxt context.Context, tx *ent.Tx, s *State) error
}

func NewState(stateID int, metadataID remote.ConnMetadataID, user stateUserAccessor, delimiter string) *State {
	return &State{
		user:         user,
		stateID:      stateID,
		metadataID:   metadataID,
		doneCh:       make(chan struct{}),
		snap:         nil,
		delimiter:    delimiter,
		updatesQueue: queue.NewQueuedChannel[stateUpdate](32, 128),
	}
}

func (state *State) UserID() string {
	return state.user.getUserID()
}

func (state *State) db() *DB {
	return state.user.getDB()
}

func (state *State) List(ctx context.Context, ref, pattern string, subscribed bool, fn func(map[string]Match) error) error {
	return state.db().Read(ctx, func(ctx context.Context, client *ent.Client) error {
		mailboxes, err := DBGetAllMailboxes(ctx, client)
		if err != nil {
			return err
		}

		matches, err := getMatches(ctx, client, mailboxes, ref, pattern, state.delimiter, subscribed)
		if err != nil {
			return err
		}

		return fn(matches)
	})
}

func (state *State) Select(ctx context.Context, name string, fn func(*Mailbox) error) error {
	mbox, err := DBReadResult(ctx, state.db(), func(ctx context.Context, client *ent.Client) (*ent.Mailbox, error) {
		return DBGetMailboxByName(ctx, client, name)
	})
	if err != nil {
		return ErrNoSuchMailbox
	}

	if state.snap != nil {
		if err := state.close(); err != nil {
			return err
		}
	}

	snap, err := DBReadResult(ctx, state.db(), func(ctx context.Context, client *ent.Client) (*snapshot, error) {
		return newSnapshot(ctx, state, client, mbox)
	})
	if err != nil {
		return err
	}

	if err := state.db().Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		return DBClearRecentFlags(ctx, tx, mbox.MailboxID)
	}); err != nil {
		return err
	}

	state.snap = snap
	state.ro = false

	return fn(newMailbox(mbox, state, state.snap))
}

func (state *State) Examine(ctx context.Context, name string, fn func(*Mailbox) error) error {
	mbox, err := DBReadResult(ctx, state.db(), func(ctx context.Context, client *ent.Client) (*ent.Mailbox, error) {
		return DBGetMailboxByName(ctx, client, name)
	})
	if err != nil {
		return ErrNoSuchMailbox
	}

	if state.snap != nil {
		if err := state.close(); err != nil {
			return err
		}
	}

	snap, err := DBReadResult(ctx, state.db(), func(ctx context.Context, client *ent.Client) (*snapshot, error) {
		return newSnapshot(ctx, state, client, mbox)
	})
	if err != nil {
		return err
	}

	state.snap = snap
	state.ro = true

	return fn(newMailbox(mbox, state, state.snap))
}

func (state *State) Create(ctx context.Context, name string) error {
	mboxesToCreate, err := DBReadResult(ctx, state.db(), func(ctx context.Context, client *ent.Client) ([]string, error) {
		var mboxesToCreate []string
		// If the mailbox name is suffixed with the server's hierarchy separator, remove the separator and still create
		// the mailbox
		if strings.HasSuffix(name, state.delimiter) {
			name = strings.TrimRight(name, state.delimiter)
		}

		if exists, err := DBMailboxExistsWithName(ctx, client, name); err != nil {
			return nil, err
		} else if exists {
			return nil, ErrExistingMailbox
		}

		for _, superior := range listSuperiors(name, state.delimiter) {
			if exists, err := DBMailboxExistsWithName(ctx, client, superior); err != nil {
				return nil, err
			} else if exists {
				continue
			}

			mboxesToCreate = append(mboxesToCreate, superior)
		}

		mboxesToCreate = append(mboxesToCreate, name)

		return mboxesToCreate, nil
	})
	if err != nil {
		return err
	}

	return state.db().Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		for _, mboxName := range mboxesToCreate {
			if err := state.actionCreateMailbox(ctx, tx, mboxName); err != nil {
				return err
			}
		}

		return nil
	})
}

func (state *State) Delete(ctx context.Context, name string) error {
	mbox, err := DBReadResult(ctx, state.db(), func(ctx context.Context, client *ent.Client) (*ent.Mailbox, error) {
		return DBGetMailboxByName(ctx, client, name)
	})
	if err != nil {
		return ErrNoSuchMailbox
	}

	return state.db().Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		return state.actionDeleteMailbox(ctx, tx, mbox.RemoteID)
	})
}

func (state *State) Rename(ctx context.Context, oldName, newName string) error {
	type Result struct {
		MBox           *ent.Mailbox
		MBoxesToCreate []string
	}

	result, err := DBReadResult(ctx, state.db(), func(ctx context.Context, client *ent.Client) (Result, error) {
		mbox, err := DBGetMailboxByName(ctx, client, oldName)
		if err != nil {
			return Result{}, ErrNoSuchMailbox
		}

		if exists, err := DBMailboxExistsWithName(ctx, client, newName); err != nil {
			return Result{}, err
		} else if exists {
			return Result{}, ErrExistingMailbox
		}

		var mboxesToCreate []string
		for _, superior := range listSuperiors(newName, state.delimiter) {
			if exists, err := DBMailboxExistsWithName(ctx, client, superior); err != nil {
				return Result{}, err
			} else if exists {
				if superior == oldName {
					return Result{}, ErrExistingMailbox
				}
				continue
			}

			mboxesToCreate = append(mboxesToCreate, superior)
		}

		return Result{
			MBox:           mbox,
			MBoxesToCreate: mboxesToCreate,
		}, nil
	})
	if err != nil {
		return err
	}

	return state.db().Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		for _, m := range result.MBoxesToCreate {
			internalID, res, err := state.user.getRemote().CreateMailbox(ctx, state.metadataID, strings.Split(m, state.delimiter))
			if err != nil {
				return err
			}

			if err := DBCreateMailboxIfNotExists(ctx, tx, internalID, res, state.delimiter); err != nil {
				return err
			}
		}

		if oldName == imap.Inbox {
			return state.renameInbox(ctx, tx, result.MBox, newName)
		}

		mailboxes, err := DBGetAllMailboxes(ctx, tx.Client())
		if err != nil {
			return err
		}

		inferiors := listInferiors(oldName, state.delimiter, xslices.Map(mailboxes, func(mailbox *ent.Mailbox) string {
			return mailbox.Name
		}))

		for _, inferior := range inferiors {
			mbox, err := DBGetMailboxByName(ctx, tx.Client(), inferior)
			if err != nil {
				return ErrNoSuchMailbox
			}

			newInferior := newName + strings.TrimPrefix(inferior, oldName)

			if err := state.actionUpdateMailbox(ctx, tx, mbox.RemoteID, inferior, newInferior); err != nil {
				return err
			}
		}

		return state.actionUpdateMailbox(ctx, tx, result.MBox.RemoteID, oldName, newName)
	})
}

func (state *State) Subscribe(ctx context.Context, name string) error {
	mbox, err := DBReadResult(ctx, state.db(), func(ctx context.Context, client *ent.Client) (*ent.Mailbox, error) {
		return DBGetMailboxByName(ctx, client, name)
	})
	if err != nil {
		return ErrNoSuchMailbox
	} else if mbox.Subscribed {
		return ErrAlreadySubscribed
	}

	return state.db().Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		return mbox.Update().SetSubscribed(true).Exec(ctx)
	})
}

func (state *State) Unsubscribe(ctx context.Context, name string) error {
	mbox, err := DBReadResult(ctx, state.db(), func(ctx context.Context, client *ent.Client) (*ent.Mailbox, error) {
		return DBGetMailboxByName(ctx, client, name)
	})
	if err != nil {
		return ErrNoSuchMailbox
	} else if !mbox.Subscribed {
		return ErrAlreadyUnsubscribed
	}

	return state.db().Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
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
	mbox, err := DBReadResult(ctx, state.db(), func(ctx context.Context, client *ent.Client) (*ent.Mailbox, error) {
		return DBGetMailboxByName(ctx, client, name)
	})
	if err != nil {
		return ErrNoSuchMailbox
	}

	if state.snap != nil && state.snap.mboxID.InternalID == mbox.MailboxID {
		return fn(newMailbox(mbox, state, state.snap))
	}

	snap, err := DBReadResult(ctx, state.db(), func(ctx context.Context, client *ent.Client) (*snapshot, error) {
		return newSnapshot(ctx, state, client, mbox)
	})
	if err != nil {
		return err
	}

	return fn(newMailbox(mbox, state, snap))
}

func (state *State) Selected(ctx context.Context, fn func(*Mailbox) error) error {
	if !state.IsSelected() {
		return ErrSessionNotSelected
	}

	mbox, err := DBReadResult(ctx, state.db(), func(ctx context.Context, client *ent.Client) (*ent.Mailbox, error) {
		return DBGetMailboxByID(ctx, client, state.snap.mboxID.InternalID)
	})

	if err != nil {
		return ErrNoSuchMailbox
	}

	return fn(newMailbox(mbox, state, state.snap))
}

func (state *State) IsSelected() bool {
	return state.snap != nil
}

func (state *State) SetConnMetadataKeyValue(key string, value any) error {
	return state.user.getRemote().SetConnMetadataValue(state.metadataID, key, value)
}

func (state *State) Done() <-chan struct{} {
	if state == nil {
		return nil
	}

	return state.doneCh
}

func (state *State) Close(ctx context.Context) error {
	return state.user.removeState(ctx, state.stateID)
}

func (state *State) GetStateUpdatesCh() <-chan stateUpdate {
	if state == nil {
		return nil
	}

	return state.updatesQueue.GetChannel()
}

func (state *State) queueUpdates(updates ...stateUpdate) bool {
	return state.updatesQueue.Queue(updates...)
}

func (state *State) ApplyUpdate(ctx context.Context, update stateUpdate) error {
	logrus.WithField("StateUpdate", update).Tracef("Applying state update on state %v", state.stateID)

	if !update.filter(state) {
		return nil
	}

	return state.db().Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		return update.apply(ctx, tx, state)
	})
}

// renameInbox creates a new mailbox and moves everything there.
func (state *State) renameInbox(ctx context.Context, tx *ent.Tx, inbox *ent.Mailbox, newName string) error {
	mbox, err := state.actionCreateAndGetMailbox(ctx, tx, newName)
	if err != nil {
		return err
	}

	messages, err := DBGetMailboxMessages(ctx, tx.Client(), inbox)
	if err != nil {
		return err
	}

	messageIDs := xslices.Map(messages, func(messageUID *ent.UID) MessageIDPair {
		return NewMessageIDPair(messageUID.Edges.Message)
	})

	mboxIDPair := NewMailboxIDPair(mbox)

	if _, err := state.actionMoveMessages(ctx, tx, messageIDs, NewMailboxIDPair(inbox), mboxIDPair); err != nil {
		return err
	}

	return nil
}

func (state *State) beginIdle(ctx context.Context) ([]response.Response, error) {
	var res []response.Response

	var err error

	if res, err = DBWriteResult(ctx, state.db(), func(ctx context.Context, tx *ent.Tx) ([]response.Response, error) {
		return state.flushResponses(ctx, tx, true)
	}); err != nil {
		return nil, err
	}

	state.idleCh = make(chan response.Response)

	return res, nil
}

func (state *State) endIdle() {
	close(state.idleCh)

	state.idleCh = nil
}

func (state *State) getLiteral(messageID imap.InternalMessageID) ([]byte, error) {
	return state.user.getStore().Get(messageID)
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
	if state.idleCh == nil {
		return state.queueResponder(responder...)
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

func (state *State) queueResponder(responder ...responder) error {
	state.res = append(state.res, responder...)

	return nil
}

// popResponders pops responders from the state.
// If permitExpunge is false, it stops when it encounters the first expunge responder.
// This is important: suppose someone first expunges a message (generating an EXPUNGE)
// then puts it back (generating an EXISTS). If we didn't stop at the first expunge,
// we would send the EXISTS to the client, followed by an expunge afterwards (wrong).
func (state *State) popResponders(permitExpunge bool) []responder {
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
	if err := state.user.getRemote().DeleteConnMetadataStore(state.metadataID); err != nil {
		return fmt.Errorf("failed to delete conn metadata store: %w", err)
	}

	return nil
}

func (state *State) closeUpdateQueue() {
	state.updatesQueue.Close()
}

func (state *State) close() error {
	state.snap = nil

	state.res = nil

	return nil
}
