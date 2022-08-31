package state

import (
	"context"

	"github.com/ProtonMail/gluon/imap"
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

	ApplyMessagesAddedToMailbox(
		ctx context.Context,
		tx *ent.Tx,
		mboxID imap.InternalMailboxID,
		messageIDs []imap.InternalMessageID,
	) (map[imap.InternalMessageID]int, error)

	ApplyMessagesRemovedFromMailbox(ctx context.Context,
		tx *ent.Tx,
		mboxID imap.InternalMailboxID,
		messageIDs []imap.InternalMessageID,
	) error

	ApplyMessagesMovedFromMailbox(
		ctx context.Context,
		tx *ent.Tx,
		mboxFromID, mboxToID imap.InternalMailboxID,
		messageIDs []imap.InternalMessageID,
	) (map[imap.InternalMessageID]int, error)

	QueueOrApplyStateUpdate(ctx context.Context, tx *ent.Tx, update Update) error

	ReleaseState(ctx context.Context, st *State) error
}
