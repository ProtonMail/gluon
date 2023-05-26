package state

import (
	"context"

	"github.com/ProtonMail/gluon/db"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/utils"
	"github.com/ProtonMail/gluon/store"
)

// UserInterface represents the expected behaviour for interacting with a remote user.
// Sadly, due to Go's cyclic dependencies, this needs to be an interface. The implementation of this interface
// is available in the backend package.
type UserInterface interface {
	GetUserID() string

	GetDelimiter() string

	GetDB() db.Client

	GetRemote() Connector

	GetStore() *store.WriteControlledStore

	QueueOrApplyStateUpdate(ctx context.Context, tx db.Transaction, update ...Update) error

	ReleaseState(ctx context.Context, st *State) error

	GetRecoveryMailboxID() db.MailboxIDPair

	GenerateUIDValidity() (imap.UID, error)

	GetRecoveredMessageHashesMap() *utils.MessageHashesMap
}
