package remote

import (
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/google/uuid"
)

// CreateMessage appends a message literal to the mailbox with the given ID.
func (user *User) CreateMessage(metadataID ConnMetadataID, mboxID imap.LabelID, literal []byte, flags imap.FlagSet, date time.Time) (imap.InternalMessageID, imap.Message, error) {
	internalID := imap.InternalMessageID(uuid.NewString())

	if err := user.pushOp(&OpMessageCreate{
		OperationBase: OperationBase{MetadataID: metadataID},
		InternalID:    internalID,
		MBoxID:        mboxID,
		Literal:       literal,
		Flags:         flags,
		Date:          date,
	}); err != nil {
		return "", imap.Message{}, err
	}

	return internalID, imap.Message{
		Flags: flags,
		Date:  date,
	}, nil
}

// AddMessageToMailbox adds the message with the given ID to the mailbox with the given ID.
func (user *User) AddMessagesToMailbox(metadataID ConnMetadataID, messageIDs []imap.MessageID, mboxID imap.LabelID) error {
	return user.pushOp(&OpMessageAdd{
		OperationBase: OperationBase{MetadataID: metadataID},
		MessageIDs:    messageIDs,
		MBoxID:        mboxID,
	})
}

// RemoveMessageFromMailbox removes the message with the given ID from the mailbox with the given ID.
func (user *User) RemoveMessagesFromMailbox(metadataID ConnMetadataID, messageIDs []imap.MessageID, mboxID imap.LabelID) error {
	return user.pushOp(&OpMessageRemove{
		OperationBase: OperationBase{MetadataID: metadataID},
		MessageIDs:    messageIDs,
		MBoxID:        mboxID,
	})
}

// SetMessageSeen marks the message with the given ID as seen or unseen.
func (user *User) SetMessagesSeen(metadataID ConnMetadataID, messageIDs []imap.MessageID, seen bool) error {
	return user.pushOp(&OpMessageSeen{
		OperationBase: OperationBase{MetadataID: metadataID},
		MessageIDs:    messageIDs,
		Seen:          seen,
	})
}

// SetMessageFlagged marks the message with the given ID as seen or unseen.
func (user *User) SetMessagesFlagged(metadataID ConnMetadataID, messageIDs []imap.MessageID, flagged bool) error {
	return user.pushOp(&OpMessageFlagged{
		OperationBase: OperationBase{MetadataID: metadataID},
		MessageIDs:    messageIDs,
		Flagged:       flagged,
	})
}
