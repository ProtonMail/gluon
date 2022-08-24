package remote

import (
	"context"
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/google/uuid"
)

// CreateMessage appends a message literal to the mailbox with the given ID.
func (user *User) CreateMessage(
	ctx context.Context,
	metadataID ConnMetadataID,
	mboxID imap.LabelID,
	literal []byte,
	flags imap.FlagSet,
	date time.Time,
) (imap.InternalMessageID, imap.Message, error) {
	ctx = user.newContextWithIMAPID(ctx, metadataID)

	msg, err := user.conn.CreateMessage(ctx, mboxID, literal, flags, date)
	if err != nil {
		return "", imap.Message{}, err
	}

	internalID := imap.InternalMessageID(uuid.NewString())

	return internalID, msg, nil
}

// AddMessagesToMailbox adds the message with the given ID to the mailbox with the given ID.
func (user *User) AddMessagesToMailbox(
	ctx context.Context,
	metadataID ConnMetadataID,
	messageIDs []imap.MessageID,
	mboxID imap.LabelID,
) error {
	ctx = user.newContextWithIMAPID(ctx, metadataID)

	if err := user.conn.LabelMessages(ctx, messageIDs, mboxID); err != nil {
		return user.refresh(ctx, messageIDs, mboxID)
	}

	return nil
}

// RemoveMessagesFromMailbox removes the message with the given ID from the mailbox with the given ID.
func (user *User) RemoveMessagesFromMailbox(
	ctx context.Context,
	metadataID ConnMetadataID,
	messageIDs []imap.MessageID,
	mboxID imap.LabelID,
) error {
	ctx = user.newContextWithIMAPID(ctx, metadataID)

	if err := user.conn.UnlabelMessages(ctx, messageIDs, mboxID); err != nil {
		return user.refresh(ctx, messageIDs, mboxID)
	}

	return nil
}

// MoveMessagesFromMailbox removes the message with the given ID from the mailbox with the given ID.
func (user *User) MoveMessagesFromMailbox(
	ctx context.Context,
	metadataID ConnMetadataID,
	messageIDs []imap.MessageID,
	mboxFromID imap.LabelID,
	mboxToID imap.LabelID,
) error {
	ctx = user.newContextWithIMAPID(ctx, metadataID)

	if err := user.conn.MoveMessages(ctx, messageIDs, mboxFromID, mboxToID); err != nil {
		return user.refresh(ctx, messageIDs, mboxFromID)
	}

	return nil
}

// SetMessagesSeen marks the message with the given ID as seen or unseen.
func (user *User) SetMessagesSeen(ctx context.Context, metadataID ConnMetadataID, messageIDs []imap.MessageID, seen bool) error {
	ctx = user.newContextWithIMAPID(ctx, metadataID)

	if err := user.conn.MarkMessagesSeen(ctx, messageIDs, seen); err != nil {
		return user.refresh(ctx, messageIDs)
	}

	return nil
}

// SetMessagesFlagged marks the message with the given ID as seen or unseen.
func (user *User) SetMessagesFlagged(ctx context.Context, metadataID ConnMetadataID, messageIDs []imap.MessageID, flagged bool) error {
	ctx = user.newContextWithIMAPID(ctx, metadataID)

	if err := user.conn.MarkMessagesFlagged(ctx, messageIDs, flagged); err != nil {
		return user.refresh(ctx, messageIDs)
	}

	return nil
}
