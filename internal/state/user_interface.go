package state

import (
	"context"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db"
	"github.com/ProtonMail/gluon/internal/db/ent"
	"github.com/ProtonMail/gluon/store"
)

// UserInterface represents the expected behaviour for interacting with a remote user.
// Sadly, due to Go's cyclic dependencies, this needs to be an interface. The implementation of this interface
// is available in the backend package.
type UserInterface interface {
	GetUserID() string

	GetDelimiter() string

	GetDB() *db.DB

	GetRemote() Connector

	GetStore() store.Store

	QueueOrApplyStateUpdate(ctx context.Context, tx *ent.Tx, update Update) error

	ReleaseState(ctx context.Context, st *State) error

	GetGlobalUIDValidity() imap.UID

	SetGlobalUIDValidity(imap.UID)
}
