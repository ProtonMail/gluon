package state

import (
	"context"
	"time"

	"github.com/ProtonMail/gluon/imap"
)

type Connector interface {
	// SetConnMetadataValue sets a metadata value associated with the current connector.
	SetConnMetadataValue(key string, value any)

	// ClearConnMetadataValue clears a metadata value associated with the current connector.
	ClearConnMetadataValue(key string)

	// ClearAllConnMetadata clears all metadata values associated with the current connector.
	ClearAllConnMetadata()

	// CreateMailbox creates a new mailbox with the given name.
	CreateMailbox(ctx context.Context, name []string) (imap.InternalMailboxID, imap.Mailbox, error)

	// UpdateMailbox sets the name of the mailbox with the given ID to the given new name.
	UpdateMailbox(ctx context.Context, mboxID imap.LabelID, oldName, newName []string) error

	// DeleteMailbox deletes the mailbox with the given ID and name.
	DeleteMailbox(ctx context.Context, mboxID imap.LabelID) error

	// CreateMessage appends a message literal to the mailbox with the given ID.
	CreateMessage(
		ctx context.Context,
		mboxID imap.LabelID,
		literal []byte,
		flags imap.FlagSet,
		date time.Time,
	) (imap.InternalMessageID, imap.Message, error)

	// AddMessagesToMailbox adds the message with the given ID to the mailbox with the given ID.
	AddMessagesToMailbox(
		ctx context.Context,
		messageIDs []imap.MessageID,
		mboxID imap.LabelID,
	) error

	// RemoveMessagesFromMailbox removes the message with the given ID from the mailbox with the given ID.
	RemoveMessagesFromMailbox(
		ctx context.Context,
		messageIDs []imap.MessageID,
		mboxID imap.LabelID,
	) error

	// MoveMessagesFromMailbox removes the message with the given ID from the mailbox with the given ID.
	MoveMessagesFromMailbox(
		ctx context.Context,
		messageIDs []imap.MessageID,
		mboxFromID imap.LabelID,
		mboxToID imap.LabelID,
	) error

	// SetMessagesSeen marks the message with the given ID as seen or unseen.
	SetMessagesSeen(ctx context.Context, messageIDs []imap.MessageID, seen bool) error

	// SetMessagesFlagged marks the message with the given ID as seen or unseen.
	SetMessagesFlagged(ctx context.Context, messageIDs []imap.MessageID, flagged bool) error
}
