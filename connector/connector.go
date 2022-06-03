package connector

import (
	"context"
	"time"

	"github.com/ProtonMail/gluon/imap"
)

// Connector connects the gluon server to a remote mail store.
type Connector interface {
	// Authorize returns whether the given username/password combination are valid for this connector.
	Authorize(username, password string) bool

	// GetUpdates returns a stream of updates that the gluon server should apply.
	GetUpdates() <-chan imap.Update

	// Pause pauses the stream of updates.
	Pause()

	// Resume resumes the stream of updates.
	Resume()

	// ValidateCreate checks whether a mailbox with the given name can be created.
	// If so, the flags, permanent flags and attributes which the mailbox would have are returned.
	ValidateCreate(name []string) (flags, permFlags, attrs imap.FlagSet, err error)

	// ValidateUpdate checks whether a mailbox's name can be changed from oldName to newName.
	ValidateUpdate(oldName, newName []string) error

	// ValidateDelete checks whether the mailbox with the given name can be deleted.
	ValidateDelete(name []string) error

	// GetLabel returns information about the label with the given ID.
	GetLabel(ctx context.Context, labelID string) (imap.Mailbox, error)

	// CreateLabel creates a label with the given name.
	CreateLabel(ctx context.Context, name []string) (imap.Mailbox, error)

	// UpdateLabel sets the name of the label with the given ID.
	UpdateLabel(ctx context.Context, mboxID string, newName []string) error

	// DeleteLabel deletes the label with the given ID.
	DeleteLabel(ctx context.Context, mboxID string) error

	GetMessage(ctx context.Context, messageID string) (imap.Message, []string, error)
	CreateMessage(ctx context.Context, mboxID string, literal []byte, flags imap.FlagSet, date time.Time) (imap.Message, error)
	LabelMessages(ctx context.Context, messageIDs []string, mboxID string) error
	UnlabelMessages(ctx context.Context, messageIDs []string, mboxID string) error
	MarkMessagesSeen(ctx context.Context, messageIDs []string, seen bool) error
	MarkMessagesFlagged(ctx context.Context, messageIDs []string, flagged bool) error

	Close() error
}
