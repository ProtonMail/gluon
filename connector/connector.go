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

	// ValidateCreate checks whether a mailbox with the given name can be created.
	// If so, the flags, permanent flags and attributes which the mailbox would have are returned.
	ValidateCreate(name []string) (flags, permFlags, attrs imap.FlagSet, err error)

	// ValidateUpdate checks whether a mailbox's name can be changed from oldName to newName.
	ValidateUpdate(oldName, newName []string) error

	// ValidateDelete checks whether the mailbox with the given name can be deleted.
	ValidateDelete(name []string) error

	// GetLabel returns information about the label with the given ID.
	GetLabel(ctx context.Context, labelID imap.LabelID) (imap.Mailbox, error)

	// CreateLabel creates a label with the given name.
	CreateLabel(ctx context.Context, name []string) (imap.Mailbox, error)

	// UpdateLabel sets the name of the label with the given ID.
	UpdateLabel(ctx context.Context, labelID imap.LabelID, newName []string) error

	// DeleteLabel deletes the label with the given ID.
	DeleteLabel(ctx context.Context, labelID imap.LabelID) error

	// GetMessage returns the message with the given ID.
	GetMessage(ctx context.Context, messageID imap.MessageID) (imap.Message, []imap.LabelID, error)

	// CreateMessage creates a new message on the remote.
	CreateMessage(ctx context.Context, labelID imap.LabelID, literal []byte, flags imap.FlagSet, date time.Time) (imap.Message, error)

	// LabelMessages labels the given messages with the given label ID.
	LabelMessages(ctx context.Context, messageIDs []imap.MessageID, labelID imap.LabelID) error

	// UnlabelMessages unlabels the given messages with the given label ID.
	UnlabelMessages(ctx context.Context, messageIDs []imap.MessageID, labelID imap.LabelID) error

	MoveMessages(ctx context.Context, messageIDs []imap.MessageID, labelFromID, labelToID imap.LabelID) error

	// MarkMessagesSeen sets the seen value of the given messages.
	MarkMessagesSeen(ctx context.Context, messageIDs []imap.MessageID, seen bool) error

	// MarkMessagesFlagged sets the flagged value of the given messages.
	MarkMessagesFlagged(ctx context.Context, messageIDs []imap.MessageID, flagged bool) error

	// GetUpdates returns a stream of updates that the gluon server should apply.
	GetUpdates() <-chan imap.Update

	// Pause pauses the stream of updates.
	Pause()

	// Resume resumes the stream of updates.
	Resume()

	// Close the connector will no longer be used and all resources should be closed/released.
	Close(ctx context.Context) error
}
