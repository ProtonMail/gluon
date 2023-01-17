package state

import (
	"context"
	"time"

	"github.com/ProtonMail/gluon/imap"
)

// Connector interface for State differs slightly from the connector.Connector interface as it needs the ability
// to generate internal IDs for each request as well as track local metadata associated with each state. The local
// metadata (e.g.: IMAP ID extension info) should be injected into the context before each call to ensure the
// connector.Connector can receive this information.
// Sadly, due to Go's cyclic dependencies, this needs to be an interface. The implementation of this interface
// is available in the backend package.
type Connector interface {
	// SetConnMetadataValue sets a metadata value associated with the current connector.
	SetConnMetadataValue(key string, value any)

	// ClearConnMetadataValue clears a metadata value associated with the current connector.
	ClearConnMetadataValue(key string)

	// ClearAllConnMetadata clears all metadata values associated with the current connector.
	ClearAllConnMetadata()

	// CreateMailbox creates a new mailbox with the given name.
	CreateMailbox(ctx context.Context, name []string) (imap.Mailbox, error)

	// UpdateMailbox sets the name of the mailbox with the given ID to the given new name.
	UpdateMailbox(ctx context.Context, mboxID imap.MailboxID, newName []string) error

	// DeleteMailbox deletes the mailbox with the given ID and name.
	DeleteMailbox(ctx context.Context, mboxID imap.MailboxID) error

	// CreateMessage appends a message literal to the mailbox with the given ID.
	CreateMessage(
		ctx context.Context,
		mboxID imap.MailboxID,
		literal []byte,
		flags imap.FlagSet,
		date time.Time,
	) (imap.InternalMessageID, imap.Message, []byte, error)

	// GetMessageLiteral retrieves the message literal from the connector.
	// Note: this can get called from different go routines.
	GetMessageLiteral(ctx context.Context, id imap.MessageID) ([]byte, error)

	// AddMessagesToMailbox adds the message with the given ID to the mailbox with the given ID.
	AddMessagesToMailbox(
		ctx context.Context,
		messageIDs []imap.MessageID,
		mboxID imap.MailboxID,
	) error

	// RemoveMessagesFromMailbox removes the message with the given ID from the mailbox with the given ID.
	RemoveMessagesFromMailbox(
		ctx context.Context,
		messageIDs []imap.MessageID,
		mboxID imap.MailboxID,
	) error

	// MoveMessagesFromMailbox removes the message with the given ID from the mailbox with the given ID.
	MoveMessagesFromMailbox(
		ctx context.Context,
		messageIDs []imap.MessageID,
		mboxFromID imap.MailboxID,
		mboxToID imap.MailboxID,
	) (bool, error)

	// SetMessagesSeen marks the message with the given ID as seen or unseen.
	SetMessagesSeen(ctx context.Context, messageIDs []imap.MessageID, seen bool) error

	// SetMessagesFlagged marks the message with the given ID as seen or unseen.
	SetMessagesFlagged(ctx context.Context, messageIDs []imap.MessageID, flagged bool) error

	// SetUIDValidity sets the UID Validity for the user.
	SetUIDValidity(imap.UID) error

	// IsMailboxVisible checks whether a mailbox is visible to a client.
	IsMailboxVisible(ctx context.Context, id imap.MailboxID) bool
}
