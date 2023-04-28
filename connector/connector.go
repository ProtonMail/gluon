package connector

import (
	"context"
	"errors"
	"time"

	"github.com/ProtonMail/gluon/imap"
)

var ErrOperationNotAllowed = errors.New("operation not allowed")
var ErrMessageSizeExceedsLimits = errors.New("message size exceeds limits")

// Connector connects the gluon server to a remote mail store.
type Connector interface {
	// Authorize returns whether the given username/password combination are valid for this connector.
	Authorize(username string, password []byte) bool

	// CreateMailbox creates a mailbox with the given name.
	CreateMailbox(ctx context.Context, name []string) (imap.Mailbox, error)

	// GetMessageLiteral is intended to be used by Gluon when, for some reason, the local cached data no longer exists.
	// Note: this can get called from different go routines.
	GetMessageLiteral(ctx context.Context, id imap.MessageID) ([]byte, error)

	// GetMailboxVisibility can be used to retrieve the visibility of mailboxes for connected clients.
	GetMailboxVisibility(ctx context.Context, mboxID imap.MailboxID) imap.MailboxVisibility

	// UpdateMailboxName sets the name of the mailbox with the given ID.
	UpdateMailboxName(ctx context.Context, mboxID imap.MailboxID, newName []string) error

	// DeleteMailbox deletes the mailbox with the given ID.
	DeleteMailbox(ctx context.Context, mboxID imap.MailboxID) error

	// CreateMessage creates a new message on the remote.
	CreateMessage(ctx context.Context, mboxID imap.MailboxID, literal []byte, flags imap.FlagSet, date time.Time) (imap.Message, []byte, error)

	// AddMessagesToMailbox adds the given messages to the given mailbox.
	AddMessagesToMailbox(ctx context.Context, messageIDs []imap.MessageID, mboxID imap.MailboxID) error

	// RemoveMessagesFromMailbox removes the given messages from the given mailbox.
	RemoveMessagesFromMailbox(ctx context.Context, messageIDs []imap.MessageID, mboxID imap.MailboxID) error

	// MoveMessages removes the given messages from one mailbox and adds them to the another mailbox.
	// Returns true if the original messages should be removed from mboxFromID (e.g: Distinguishing between labels and folders).
	MoveMessages(ctx context.Context, messageIDs []imap.MessageID, mboxFromID, mboxToID imap.MailboxID) (bool, error)

	// MarkMessagesSeen sets the seen value of the given messages.
	MarkMessagesSeen(ctx context.Context, messageIDs []imap.MessageID, seen bool) error

	// MarkMessagesFlagged sets the flagged value of the given messages.
	MarkMessagesFlagged(ctx context.Context, messageIDs []imap.MessageID, flagged bool) error

	// GetUpdates returns a stream of updates that the gluon server should apply.
	GetUpdates() <-chan imap.Update

	// Close the connector will no longer be used and all resources should be closed/released.
	Close(ctx context.Context) error
}
