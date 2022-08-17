package remote

import (
	"context"
	"github.com/ProtonMail/gluon/imap"
	"github.com/google/uuid"
)

// CreateMailbox creates a new mailbox with the given name.
func (user *User) CreateMailbox(ctx context.Context, metadataID ConnMetadataID, name []string) (imap.InternalMailboxID, imap.Mailbox, error) {
	ctx = user.newContextWithIMAPID(ctx, metadataID)

	internalID := imap.InternalMailboxID(uuid.NewString())

	mbox, err := user.conn.CreateLabel(ctx, name)
	if err != nil {
		return "", imap.Mailbox{}, err
	}

	return internalID, mbox, nil
}

// UpdateMailbox sets the name of the mailbox with the given ID to the given new name.
func (user *User) UpdateMailbox(ctx context.Context, metadataID ConnMetadataID, mboxID imap.LabelID, newName []string) error {
	ctx = user.newContextWithIMAPID(ctx, metadataID)

	return user.conn.UpdateLabel(ctx, mboxID, newName)
}

// DeleteMailbox deletes the mailbox with the given ID and name.
func (user *User) DeleteMailbox(ctx context.Context, metadataID ConnMetadataID, mboxID imap.LabelID) error {
	ctx = user.newContextWithIMAPID(ctx, metadataID)

	return user.conn.DeleteLabel(ctx, mboxID)
}
