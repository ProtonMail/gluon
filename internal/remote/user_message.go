package remote

import (
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/google/uuid"
)

// CreateMessage appends a message literal to the mailbox with the given ID.
func (user *User) CreateMessage(metadataID ConnMetadataID, mboxID string, literal []byte, flags imap.FlagSet, date time.Time) (imap.Message, error) {
	tempID := uuid.NewString()

	if err := user.pushOp(&OpMessageCreate{
		OperationBase: OperationBase{MetadataID: metadataID},
		TempID:        tempID,
		MBoxID:        mboxID,
		Literal:       literal,
		Flags:         flags,
		Date:          date,
	}); err != nil {
		return imap.Message{}, err
	}

	return imap.Message{
		ID:    tempID,
		Flags: flags,
		Date:  date,
	}, nil
}

// AddMessageToMailbox adds the message with the given ID to the mailbox with the given ID.
func (user *User) AddMessagesToMailbox(metadataID ConnMetadataID, messageIDs []string, mboxID string) error {
	return user.pushOp(&OpMessageAdd{
		OperationBase: OperationBase{MetadataID: metadataID},
		MessageIDs:    messageIDs,
		MBoxID:        mboxID,
	})
}

// RemoveMessageFromMailbox removes the message with the given ID from the mailbox with the given ID.
func (user *User) RemoveMessagesFromMailbox(metadataID ConnMetadataID, messageIDs []string, mboxID string) error {
	return user.pushOp(&OpMessageRemove{
		OperationBase: OperationBase{MetadataID: metadataID},
		MessageIDs:    messageIDs,
		MBoxID:        mboxID,
	})
}

// SetMessageSeen marks the message with the given ID as seen or unseen.
func (user *User) SetMessagesSeen(metadataID ConnMetadataID, messageIDs []string, seen bool) error {
	return user.pushOp(&OpMessageSeen{
		OperationBase: OperationBase{MetadataID: metadataID},
		MessageIDs:    messageIDs,
		Seen:          seen,
	})
}

// SetMessageFlagged marks the message with the given ID as seen or unseen.
func (user *User) SetMessagesFlagged(metadataID ConnMetadataID, messageIDs []string, flagged bool) error {
	return user.pushOp(&OpMessageFlagged{
		OperationBase: OperationBase{MetadataID: metadataID},
		MessageIDs:    messageIDs,
		Flagged:       flagged,
	})
}
