package remote

import (
	"github.com/ProtonMail/gluon/imap"
	"github.com/google/uuid"
)

// CreateMailbox creates a new mailbox with the given name.
func (user *User) CreateMailbox(metadataID ConnMetadataID, name []string) (imap.Mailbox, error) {
	flags, permFlags, attrs, err := user.conn.ValidateCreate(name)
	if err != nil {
		return imap.Mailbox{}, err
	}

	tempID := uuid.NewString()

	if err := user.pushOp(&OpMailboxCreate{
		OperationBase: OperationBase{MetadataID: metadataID},
		TempID:        tempID,
		Name:          name,
	}); err != nil {
		return imap.Mailbox{}, err
	}

	return imap.Mailbox{
		ID:   tempID,
		Name: name,

		Flags:          flags,
		PermanentFlags: permFlags,
		Attributes:     attrs,
	}, nil
}

// UpdateMailbox sets the name of the mailbox with the given ID to the given new name.
func (user *User) UpdateMailbox(metadataID ConnMetadataID, mboxID string, oldName, newName []string) error {
	if err := user.conn.ValidateUpdate(oldName, newName); err != nil {
		return err
	}

	return user.pushOp(&OpMailboxUpdate{
		OperationBase: OperationBase{MetadataID: metadataID},
		MBoxID:        mboxID,
		Name:          newName,
	})
}

// DeleteMailbox deletes the mailbox with the given ID and name.
func (user *User) DeleteMailbox(metadataID ConnMetadataID, mboxID string, name []string) error {
	if err := user.conn.ValidateDelete(name); err != nil {
		return err
	}

	return user.pushOp(&OpMailboxDelete{
		OperationBase: OperationBase{MetadataID: metadataID},
		MBoxID:        mboxID,
	})
}
