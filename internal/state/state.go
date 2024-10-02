package state

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/gluon/db"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/ids"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/limits"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/gluon/rfc822"
	"github.com/bradenaw/juniper/xmaps"
	"github.com/bradenaw/juniper/xslices"
	"github.com/sirupsen/logrus"
)

type StateID int64

// State represents the active session's state after a user has been authenticated. This code is expected to run on
// one single goroutine/thread and should some interaction be required with other states from other session, they shall
// happen through updates (state.Update) which are queued via the UserInterface.
// Similarly, the state also accepts incoming updates via the ApplyUpdate method.
type State struct {
	user UserInterface

	StateID StateID

	idleCh chan response.Response

	res []Responder

	snap *snapshot
	ro   bool

	doneCh chan struct{}

	updatesQueue *async.QueuedChannel[Update]

	delimiter string

	// invalid indicates whether this state became invalid and a clients needs to disconnect.
	invalid bool

	imapLimits limits.IMAP

	panicHandler async.PanicHandler

	log *logrus.Entry
}

var stateIDGenerator int64

func nextStateID() StateID {
	return StateID(atomic.AddInt64(&stateIDGenerator, 1))
}

func NewState(user UserInterface, delimiter string, imapLimits limits.IMAP, panicHandler async.PanicHandler) *State {
	stateID := nextStateID()

	return &State{
		user:         user,
		StateID:      stateID,
		doneCh:       make(chan struct{}),
		snap:         nil,
		delimiter:    delimiter,
		updatesQueue: async.NewQueuedChannel[Update](32, 128, panicHandler, fmt.Sprintf("gluon-state-%v", stateID)),
		imapLimits:   imapLimits,
		panicHandler: panicHandler,
		log:          logrus.WithField("pkg", "gluon/state").WithField("state", stateID),
	}
}

func (state *State) UserID() string {
	return state.user.GetUserID()
}

func (state *State) List(ctx context.Context, ref, pattern string, lsub bool, fn func(map[string]Match) error) error {
	return stateDBRead(ctx, state, func(ctx context.Context, client db.ReadOnly) error {
		mailboxes, err := client.GetAllMailboxesWithAttr(ctx)
		if err != nil {
			return err
		}

		recoveryMailboxID := state.user.GetRecoveryMailboxID().InternalID
		recoveryMBoxMessageCount, err := client.GetMailboxMessageCount(ctx, recoveryMailboxID)
		if err != nil {
			state.log.WithError(err).Error("Failed to get recovery mailbox message count, assuming empty")
			recoveryMBoxMessageCount = 0
		}

		mailboxes = xslices.Filter(mailboxes, func(mailbox *db.MailboxWithAttr) bool {
			if mailbox.ID == recoveryMailboxID && recoveryMBoxMessageCount == 0 {
				return false
			}

			switch visibility := state.user.GetRemote().GetMailboxVisibility(ctx, mailbox.RemoteID); visibility {
			case imap.Hidden:
				return false
			case imap.Visible:
				return true
			case imap.HiddenIfEmpty:
				count, err := client.GetMailboxMessageCount(ctx, mailbox.ID)
				if err != nil {
					state.log.WithError(err).Error("Failed to get recovery mailbox message count, assuming not empty")
					return true
				}
				return count > 0
			default:
				state.log.Errorf("Unknown IMAP Mailbox visibility %v", visibility)
				return true
			}
		})

		var deletedSubscriptions map[imap.MailboxID]*db.DeletedSubscription

		if lsub {
			deletedSubscriptions, err = client.GetDeletedSubscriptionSet(ctx)
			if err != nil {
				return err
			}
		}

		// Convert existing mailboxes over to match format.
		matchMailboxes := make([]matchMailbox, 0, len(mailboxes))
		for _, mbox := range mailboxes {
			delete(deletedSubscriptions, mbox.RemoteID)

			// Only include subscribed mailboxes when LSUB is used.
			if lsub && !mbox.Subscribed {
				continue
			}

			matchMailboxes = append(matchMailboxes, matchMailbox{
				Name:       mbox.Name,
				Subscribed: lsub,
				EntMBox:    mbox,
			})
		}

		if lsub {
			// Insert any remaining mailboxes that have been deleted but are still subscribed.
			for _, s := range deletedSubscriptions {
				if state.user.GetRemote().GetMailboxVisibility(ctx, s.RemoteID) != imap.Visible {
					continue
				}

				matchMailboxes = append(matchMailboxes, matchMailbox{
					Name:       s.Name,
					Subscribed: true,
					EntMBox:    nil,
				})
			}
		}

		matches, err := getMatches(ctx, client, matchMailboxes, ref, pattern, state.delimiter, lsub)
		if err != nil {
			return err
		}

		return fn(matches)
	})
}

func (state *State) Select(ctx context.Context, name string, fn func(*Mailbox) error) error {
	mbox, err := stateDBReadResult(ctx, state, func(ctx context.Context, client db.ReadOnly) (*db.Mailbox, error) {
		return client.GetMailboxByName(ctx, name)
	})
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return ErrNoSuchMailbox
		}

		return err
	}

	if state.snap != nil {
		if err := state.close(); err != nil {
			return err
		}
	}

	snap, err := stateDBReadResult(ctx, state, func(ctx context.Context, client db.ReadOnly) (*snapshot, error) {
		return newSnapshot(ctx, state, client, mbox)
	})
	if err != nil {
		return err
	}

	if err := stateDBWrite(ctx, state, func(ctx context.Context, tx db.Transaction) ([]Update, error) {
		return nil, tx.ClearRecentFlagsInMailbox(ctx, mbox.ID)
	}); err != nil {
		return err
	}

	state.snap = snap
	state.ro = false

	return fn(newMailbox(mbox, state, state.snap))
}

func (state *State) Examine(ctx context.Context, name string, fn func(*Mailbox) error) error {
	mbox, err := stateDBReadResult(ctx, state, func(ctx context.Context, client db.ReadOnly) (*db.Mailbox, error) {
		return client.GetMailboxByName(ctx, name)
	})
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return ErrNoSuchMailbox
		}

		return err
	}

	if state.snap != nil {
		if err := state.close(); err != nil {
			return err
		}
	}

	snap, err := stateDBReadResult(ctx, state, func(ctx context.Context, client db.ReadOnly) (*snapshot, error) {
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
	uidValidity, err := state.user.GenerateUIDValidity()
	if err != nil {
		return err
	}

	if err := state.imapLimits.CheckUIDValidity(uidValidity); err != nil {
		return err
	}

	if strings.HasPrefix(strings.ToLower(name), ids.GluonRecoveryMailboxNameLowerCase) {
		return ErrOperationNotAllowed
	}

	if state.delimiter != "" {
		if strings.HasPrefix(name, state.delimiter) {
			return ErrMailboxNameBeginsWithSeparator
		}

		if strings.Contains(name, state.delimiter+state.delimiter) {
			return ErrMailboxNameAdjacentSeparator
		}
	}

	return stateDBWrite(ctx, state, func(ctx context.Context, tx db.Transaction) ([]Update, error) {
		if mailboxCount, err := tx.GetMailboxCount(ctx); err != nil {
			return nil, err
		} else if err := state.imapLimits.CheckMailBoxCount(mailboxCount); err != nil {
			return nil, err
		}

		var mboxesToCreate []string
		// If the mailbox name is suffixed with the server's hierarchy separator, remove the separator and still create
		// the mailbox
		if strings.HasSuffix(name, state.delimiter) {
			name = strings.TrimRight(name, state.delimiter)
		}

		if exists, err := tx.MailboxExistsWithName(ctx, name); err != nil {
			return nil, err
		} else if exists {
			return nil, ErrExistingMailbox
		}

		for _, superior := range listSuperiors(name, state.delimiter) {
			if exists, err := tx.MailboxExistsWithName(ctx, superior); err != nil {
				return nil, err
			} else if exists {
				continue
			}

			mboxesToCreate = append(mboxesToCreate, superior)
		}

		mboxesToCreate = append(mboxesToCreate, name)

		var allUpdates []Update

		for _, mboxName := range mboxesToCreate {
			updates, err := state.actionCreateMailbox(ctx, tx, mboxName, uidValidity)
			if err != nil {
				return nil, err
			}

			allUpdates = append(allUpdates, updates...)
		}

		return allUpdates, nil
	})
}

// Delete returns true if the mailbox that was deleted was the same as the one that was currently selected.
func (state *State) Delete(ctx context.Context, name string) (bool, error) {
	if strings.EqualFold(name, ids.GluonRecoveryMailboxName) {
		return false, ErrOperationNotAllowed
	}

	mboxID, err := stateDBWriteResult(ctx, state, func(ctx context.Context, tx db.Transaction) ([]Update, imap.InternalMailboxID, error) {
		mbox, err := tx.GetMailboxByName(ctx, name)
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				return nil, 0, ErrNoSuchMailbox
			}

			return nil, 0, err
		}

		update, err := state.actionDeleteMailbox(ctx, tx, db.NewMailboxIDPair(mbox))
		if err != nil {
			return nil, 0, err
		}

		return update, mbox.ID, nil
	})
	if err != nil {
		return false, err
	}

	return state.snap != nil && state.snap.mboxID.InternalID == mboxID, nil
}

func (state *State) Rename(ctx context.Context, oldName, newName string) error {
	type Result struct {
		MBox           *db.Mailbox
		MBoxesToCreate []string
	}

	if strings.EqualFold(oldName, ids.GluonRecoveryMailboxName) || strings.EqualFold(newName, ids.GluonRecoveryMailboxName) {
		return ErrOperationNotAllowed
	}

	return stateDBWrite(ctx, state, func(ctx context.Context, tx db.Transaction) ([]Update, error) {
		var allUpdates []Update

		mbox, err := tx.GetMailboxByName(ctx, oldName)
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				return nil, ErrNoSuchMailbox
			}

			return nil, err
		}

		if exists, err := tx.MailboxExistsWithName(ctx, newName); err != nil {
			return nil, err
		} else if exists {
			return nil, ErrExistingMailbox
		}

		var mboxesToCreate []string
		for _, superior := range listSuperiors(newName, state.delimiter) {
			if exists, err := tx.MailboxExistsWithName(ctx, superior); err != nil {
				return nil, err
			} else if exists {
				if superior == oldName {
					return nil, ErrExistingMailbox
				}
				continue
			}

			mboxesToCreate = append(mboxesToCreate, superior)
		}

		for _, m := range mboxesToCreate {
			uidValidity, err := state.user.GenerateUIDValidity()
			if err != nil {
				return nil, err
			}

			updates, res, err := state.user.GetRemote().CreateMailbox(ctx, tx, strings.Split(m, state.delimiter))
			if err != nil {
				return nil, err
			}

			allUpdates = append(allUpdates, updates...)

			if err := tx.CreateMailboxIfNotExists(ctx, res, state.delimiter, uidValidity); err != nil {
				return nil, err
			}
		}

		if oldName == imap.Inbox {
			return state.renameInbox(ctx, tx, mbox, newName)
		}

		updates, err := state.actionUpdateMailbox(ctx, tx, mbox.RemoteID, newName)
		if err != nil {
			return nil, err
		}

		allUpdates = append(allUpdates, updates...)

		// Locally update all inferiors so we don't wait for update
		mailboxes, err := tx.GetAllMailboxesWithAttr(ctx)
		if err != nil {
			return nil, err
		}

		inferiors := listInferiors(oldName, state.delimiter, xslices.Map(mailboxes, func(mailbox *db.MailboxWithAttr) string {
			return mailbox.Name
		}))

		for _, inferior := range inferiors {
			mbox, err := tx.GetMailboxByName(ctx, inferior)
			if err != nil {
				if errors.Is(err, db.ErrNotFound) {
					return nil, ErrNoSuchMailbox
				}

				return nil, err
			}

			newInferior := newName + strings.TrimPrefix(inferior, oldName)

			if err := tx.RenameMailboxWithRemoteID(ctx, mbox.RemoteID, newInferior); err != nil {
				return nil, err
			}
		}

		return allUpdates, nil
	})
}

func (state *State) Subscribe(ctx context.Context, name string) error {
	return stateDBWrite(ctx, state, func(ctx context.Context, tx db.Transaction) ([]Update, error) {
		mbox, err := tx.GetMailboxByName(ctx, name)
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				return nil, ErrNoSuchMailbox
			}

			return nil, err
		}

		if mbox.Subscribed {
			return nil, ErrAlreadySubscribed
		}

		return nil, tx.SetMailboxSubscribed(ctx, mbox.ID, true)
	})
}

func (state *State) Unsubscribe(ctx context.Context, name string) error {
	return stateDBWrite(ctx, state, func(ctx context.Context, tx db.Transaction) ([]Update, error) {
		mbox, err := tx.GetMailboxByName(ctx, name)
		if err != nil {
			// If mailbox does not exist, check that if it is present in the deleted subscription table
			if count, err := tx.RemoveDeletedSubscriptionWithName(ctx, name); err != nil {
				return nil, err
			} else if count == 0 {
				return nil, ErrNoSuchMailbox
			} else {
				return nil, nil
			}
		}

		if !mbox.Subscribed {
			return nil, ErrAlreadyUnsubscribed
		}

		return nil, tx.SetMailboxSubscribed(ctx, mbox.ID, false)
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
	mbox, err := stateDBReadResult(ctx, state, func(ctx context.Context, client db.ReadOnly) (*db.Mailbox, error) {
		return client.GetMailboxByName(ctx, name)
	})
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return ErrNoSuchMailbox
		}

		return err
	}

	if state.snap != nil && state.snap.mboxID.InternalID == mbox.ID {
		return fn(newMailbox(mbox, state, state.snap))
	}

	snap, err := stateDBReadResult(ctx, state, func(ctx context.Context, client db.ReadOnly) (*snapshot, error) {
		return newSnapshot(ctx, state, client, mbox)
	})
	if err != nil {
		return err
	}

	return fn(newMailbox(mbox, state, snap))
}

// AppendOnlyMailbox does not guarantee that the mailbox snapshot is loaded data from the database
// and passes true into the function if the currently selected mailbox matches the requested mailbox.
// It can only be used for appending.
func (state *State) AppendOnlyMailbox(ctx context.Context, name string, fn func(AppendOnlyMailbox, bool) error) error {
	if strings.EqualFold(name, ids.GluonRecoveryMailboxName) {
		return ErrOperationNotAllowed
	}

	mbox, err := stateDBReadResult(ctx, state, func(ctx context.Context, client db.ReadOnly) (*db.Mailbox, error) {
		return client.GetMailboxByName(ctx, name)
	})
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return ErrNoSuchMailbox
		}

		return err
	}

	if state.snap != nil && state.snap.mboxID.InternalID == mbox.ID {
		return fn(newMailbox(mbox, state, state.snap), true)
	}

	snap := newEmptySnapshot(state, mbox)

	return fn(newMailbox(mbox, state, snap), false)
}

func (state *State) Selected(ctx context.Context, fn func(*Mailbox) error) error {
	if !state.IsSelected() {
		return ErrSessionNotSelected
	}

	mbox, err := stateDBReadResult(ctx, state, func(ctx context.Context, client db.ReadOnly) (*db.Mailbox, error) {
		return client.GetMailboxByID(ctx, state.snap.mboxID.InternalID)
	})
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return ErrNoSuchMailbox
		}

		return err
	}

	return fn(newMailbox(mbox, state, state.snap))
}

func (state *State) IsSelected() bool {
	return state.snap != nil
}

func (state *State) SetConnMetadataKeyValue(key string, value any) {
	state.user.GetRemote().SetConnMetadataValue(key, value)
}

func (state *State) Done() <-chan struct{} {
	if state == nil {
		return nil
	}

	return state.doneCh
}

func (state *State) Close(ctx context.Context) error {
	state.deleteConnMetadata()

	state.closeUpdateQueue()

	if err := state.close(); err != nil {
		return fmt.Errorf("failed to close state: %w", err)
	}

	return nil
}

func (state *State) SignalClose() {
	close(state.doneCh)
}

func (state *State) ReleaseState(ctx context.Context) error {
	// Regrettably we can't have the session release the state during `Session.done` since it causes a deadlock when
	// removing a User.
	return state.user.ReleaseState(ctx, state)
}

func (state *State) GetStateUpdatesCh() <-chan Update {
	if state == nil {
		return nil
	}

	return state.updatesQueue.GetChannel()
}

func (state *State) QueueUpdates(updates ...Update) bool {
	return state.updatesQueue.Enqueue(updates...)
}

func (state *State) ApplyUpdate(ctx context.Context, update Update) error {
	state.log.WithField("Update", update).Debugf("Applying state update on state %v", state.StateID)

	if !update.Filter(state) {
		return nil
	}

	if err := state.user.GetDB().Write(ctx, func(ctx context.Context, tx db.Transaction) error {
		return update.Apply(ctx, tx, state)
	}); err != nil {
		// Only one has been reported in Sentry so far...
		reporter.MessageWithContext(ctx,
			"Failed to apply state update",
			reporter.Context{"error": err, "update": update.String()},
		)

		return err
	}

	return nil
}

func (state *State) HasMessage(id imap.InternalMessageID) bool {
	return state.snap != nil && state.snap.hasMessage(id)
}

func (state *State) IsValid() bool {
	return !state.invalid
}

func (state *State) markInvalid() {
	state.invalid = true
}

// renameInbox creates a new mailbox and moves everything there.
func (state *State) renameInbox(ctx context.Context, tx db.Transaction, inbox *db.Mailbox, newName string) ([]Update, error) {
	uidValidity, err := state.user.GenerateUIDValidity()
	if err != nil {
		return nil, err
	}

	var allUpdates []Update

	updates, mbox, err := state.actionCreateAndGetMailbox(ctx, tx, newName, uidValidity)
	if err != nil {
		return nil, err
	}

	allUpdates = append(allUpdates, updates...)

	messageIDs, err := tx.GetMailboxMessageIDPairs(ctx, inbox.ID)
	if err != nil {
		return nil, err
	}

	mboxIDPair := db.NewMailboxIDPair(mbox)

	updatesMove, _, err := state.actionMoveMessages(ctx, tx, messageIDs, db.NewMailboxIDPair(inbox), mboxIDPair)
	if err != nil {
		return nil, err
	}

	allUpdates = append(allUpdates, updatesMove...)

	return allUpdates, nil
}

func (state *State) beginIdle(ctx context.Context) ([]response.Response, error) {
	var res []response.Response

	res, err := state.flushResponses(ctx, true)
	if err != nil {
		return nil, err
	}

	state.idleCh = make(chan response.Response)

	return res, nil
}

func (state *State) endIdle() {
	close(state.idleCh)

	state.idleCh = nil
}

func (state *State) getLiteral(ctx context.Context, messageID db.MessageIDPair) ([]byte, error) {
	var literal []byte

	storeLiteral, firstErr := state.user.GetStore().Get(messageID.InternalID)
	if firstErr != nil {
		// Do not attempt to recovered messages from the connector.
		if ids.IsRecoveredRemoteMessageID(messageID.RemoteID) {
			state.log.Debugf("Failed load %v from store, but it is a recovered message.", messageID.InternalID)
			return nil, firstErr
		}

		state.log.Debugf("Failed load %v from store, attempting to download from connector", messageID.InternalID.ShortID())

		connectorLiteral, err := state.user.GetRemote().GetMessageLiteral(ctx, messageID.RemoteID)
		if err != nil {
			state.log.Errorf("Failed to download message from connector: %v", err)
			return nil, fmt.Errorf("message failed to load from cache (%v), failed to download from connector: %w", firstErr, err)
		}

		literalWithHeader, err := rfc822.SetHeaderValue(connectorLiteral, ids.InternalIDKey, messageID.InternalID.String())
		if err != nil {
			return nil, fmt.Errorf("failed to set internal ID on downloaded message: %w", err)
		}

		if err := state.user.GetStore().Set(messageID.InternalID, bytes.NewReader(literalWithHeader)); err != nil {
			state.log.Errorf("Failed to store download message from connector: %v", err)
			return nil, fmt.Errorf("message failed to load from cache (%v), failed to store new downloaded message: %w", firstErr, err)
		}

		state.log.Debugf("Message %v downloaded and stored ", messageID.InternalID.ShortID())

		literal = literalWithHeader
	} else {
		literal = storeLiteral
	}

	return literal, nil
}

func (state *State) flushResponses(ctx context.Context, permitExpunge bool) ([]response.Response, error) {
	var responses []response.Response

	select {
	case <-ctx.Done():
		return nil, ctx.Err()

	default: // fallthrough
	}

	var dbUpdates []responderDBUpdate

	for _, responder := range state.popResponders(permitExpunge) {
		state.log.WithField("state", state.StateID).WithField("Origin", "Flush").Debugf("Applying responder: %v", responder.String())

		res, dbUpdate, err := responder.handle(ctx, state.snap, state.StateID)
		if err != nil {
			return nil, err
		}

		responses = append(responses, res...)

		if dbUpdate != nil {
			dbUpdates = append(dbUpdates, dbUpdate)
		}
	}

	if err := state.user.GetDB().Write(ctx, func(ctx context.Context, tx db.Transaction) error {
		for _, update := range dbUpdates {
			if err := update.apply(ctx, tx); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return response.Merge(responses), nil
}

func (state *State) PushResponder(ctx context.Context, tx db.Transaction, responder ...Responder) error {
	if state.idleCh == nil {
		return state.queueResponder(responder...)
	}

	for _, responder := range responder {
		state.log.WithField("state", state.StateID).WithField("Origin", "Push").Debugf("Applying responder: %v", responder.String())

		res, dbUpdate, err := responder.handle(ctx, state.snap, state.StateID)
		if err != nil {
			return err
		}

		if dbUpdate != nil {
			if err := dbUpdate.apply(ctx, tx); err != nil {
				return err
			}
		}

		for _, res := range res {
			state.idleCh <- res
		}
	}

	return nil
}

func (state *State) queueResponder(responder ...Responder) error {
	state.res = append(state.res, responder...)

	return nil
}

// popResponders pops responders from the state.
// If permitExpunge is false, it stops when it encounters the first expunge Responder.
// This is important: suppose someone first expunges a message (generating an EXPUNGE)
// then puts it back (generating an EXISTS). If we didn't stop at the first expunge,
// we would send the EXISTS to the client, followed by an expunge afterwards (wrong).
func (state *State) popResponders(permitExpunge bool) []Responder {
	if len(state.res) == 0 {
		return nil
	}

	var pop, rem []Responder

	skipIDs := make(xmaps.Set[imap.InternalMessageID])

	for _, res := range state.res {
		if permitExpunge {
			pop = append(pop, res)
		} else if _, ok := res.(*expunge); ok {
			rem = append(rem, res)
			skipIDs.Add(res.getMessageID())
		} else if _, ok := res.(*targetedExists); ok {
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

func (state *State) UpdateMailboxRemoteID(internalID imap.InternalMailboxID, remoteID imap.MailboxID) error {
	if state.snap != nil {
		if err := state.snap.updateMailboxRemoteID(internalID, remoteID); err != nil {
			return err
		}
	}

	return nil
}

func (state *State) UpdateMessageRemoteID(internalID imap.InternalMessageID, remoteID imap.MessageID) error {
	if state.snap != nil && state.snap.hasMessage(internalID) {
		if err := state.snap.updateMessageRemoteID(internalID, remoteID); err != nil {
			return err
		}
	}

	return nil
}

// We don't want the queue closed to be reported as an error.
// User will clean up existing metadata entries by itself when closed.
func (state *State) deleteConnMetadata() {
	state.user.GetRemote().ClearAllConnMetadata()
}

func (state *State) closeUpdateQueue() {
	state.updatesQueue.Close()
}

func (state *State) close() error {
	state.snap = nil

	state.res = nil

	return nil
}

func stateDBRead(ctx context.Context, state *State, fn func(context.Context, db.ReadOnly) error) error {
	return state.user.GetDB().Read(ctx, fn)
}

func stateDBReadResult[T any](ctx context.Context, state *State, fn func(context.Context, db.ReadOnly) (T, error)) (T, error) {
	return db.ClientReadType(ctx, state.user.GetDB(), fn)
}

func stateDBWrite(ctx context.Context, state *State, fn func(context.Context, db.Transaction) ([]Update, error)) error {
	var updates []Update

	if err := state.user.GetDB().Write(ctx, func(ctx context.Context, tx db.Transaction) error {
		up, err := fn(ctx, tx)
		updates = up
		return err
	}); err != nil {
		return err
	}

	// need to create a separate transaction for the state updates so that import changes get written first.
	if len(updates) != 0 {
		if err := state.user.GetDB().Write(ctx, func(ctx context.Context, tx db.Transaction) error {
			return state.user.QueueOrApplyStateUpdate(ctx, tx, updates...)
		}); err != nil {
			return err
		}
	}

	return nil
}

func stateDBWriteResult[T any](ctx context.Context, state *State, fn func(context.Context, db.Transaction) ([]Update, T, error)) (T, error) {
	var updates []Update

	result, err := db.ClientWriteType(ctx, state.user.GetDB(), func(ctx context.Context, tx db.Transaction) (T, error) {
		up, val, err := fn(ctx, tx)
		updates = up
		return val, err
	})
	if err != nil {
		var t T
		return t, err
	}

	// need to create a separate transaction for the state updates so that import changes get written first.
	if len(updates) != 0 {
		if err := state.user.GetDB().Write(ctx, func(ctx context.Context, tx db.Transaction) error {
			return state.user.QueueOrApplyStateUpdate(ctx, tx, updates...)
		}); err != nil {
			return result, err
		}
	}

	return result, nil
}
