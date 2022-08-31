package state

import (
	"context"

	"github.com/ProtonMail/gluon/internal/db"
	"github.com/ProtonMail/gluon/internal/db/ent"
	"github.com/ProtonMail/gluon/store"
)

// UserInterface represents the expected behaviour for interacting with a remote user.
type UserInterface interface {
	GetUserID() string

	GetDelimiter() string

	GetDB() *db.DB

	GetRemote() Connector

	GetStore() store.Store

	QueueOrApplyStateUpdate(ctx context.Context, tx *ent.Tx, update Update) error

	ReleaseState(ctx context.Context, st *State) error
}
