package backend

import (
	"context"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db"
	"github.com/ProtonMail/gluon/internal/db/ent"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/ProtonMail/gluon/store"
)

// StateUserInterfaceImpl should be used to interface with the user type from a State type. This is meant to control
// the API boundary layer.
type StateUserInterfaceImpl struct {
	u *user
	c state.Connector
}

func newStateUserInterfaceImpl(u *user, connector state.Connector) *StateUserInterfaceImpl {
	return &StateUserInterfaceImpl{u: u, c: connector}
}

func (s *StateUserInterfaceImpl) GetUserID() string {
	return s.u.userID
}

func (s *StateUserInterfaceImpl) GetDelimiter() string {
	return s.u.delimiter
}

func (s *StateUserInterfaceImpl) GetDB() *db.DB {
	return s.u.db
}

func (s *StateUserInterfaceImpl) GetRemote() state.Connector {
	return s.c
}

func (s *StateUserInterfaceImpl) GetStore() store.Store {
	return s.u.store
}

func (s *StateUserInterfaceImpl) ApplyMessagesAddedToMailbox(
	ctx context.Context,
	tx *ent.Tx,
	mboxID imap.InternalMailboxID,
	messageIDs []imap.InternalMessageID,
) (map[imap.InternalMessageID]int, error) {
	return s.u.applyMessagesAddedToMailbox(ctx, tx, mboxID, messageIDs)
}

func (s *StateUserInterfaceImpl) ApplyMessagesRemovedFromMailbox(ctx context.Context,
	tx *ent.Tx,
	mboxID imap.InternalMailboxID,
	messageIDs []imap.InternalMessageID,
) error {
	return s.u.applyMessagesRemovedFromMailbox(ctx, tx, mboxID, messageIDs)
}

func (s *StateUserInterfaceImpl) ApplyMessagesMovedFromMailbox(
	ctx context.Context,
	tx *ent.Tx,
	mboxFromID, mboxToID imap.InternalMailboxID,
	messageIDs []imap.InternalMessageID,
) (map[imap.InternalMessageID]int, error) {
	return s.u.applyMessagesMovedFromMailbox(ctx, tx, mboxFromID, mboxToID, messageIDs)
}

func (s *StateUserInterfaceImpl) QueueOrApplyStateUpdate(ctx context.Context, tx *ent.Tx, update state.Update) error {
	return s.u.queueOrApplyStateUpdate(ctx, tx, update)
}

func (s *StateUserInterfaceImpl) ReleaseState(ctx context.Context, st *state.State) error {
	return s.u.removeState(ctx, st)
}
