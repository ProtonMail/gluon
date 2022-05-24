package imap

import (
	"fmt"

	"github.com/ProtonMail/gluon/internal/utils"
	"github.com/ProtonMail/gluon/rfc822"
)

type MessageCreated struct {
	*updateWaiter

	Message    Message
	Literal    []byte
	MailboxIDs []string

	Body      string
	Structure string
	Envelope  string
}

func NewMessageCreated(message Message, literal []byte, mailboxIDs []string) (MessageCreated, error) {
	root, err := rfc822.Parse(literal)
	if err != nil {
		return MessageCreated{}, fmt.Errorf("failed to parse message literal: %w", err)
	}

	body, err := Structure(root, false)
	if err != nil {
		return MessageCreated{}, fmt.Errorf("failed to build message body: %w", err)
	}

	structure, err := Structure(root, true)
	if err != nil {
		return MessageCreated{}, fmt.Errorf("failed to build message body structure: %w", err)
	}

	envelope, err := Envelope(root.ParseHeader())
	if err != nil {
		return MessageCreated{}, fmt.Errorf("failed to build message envelope: %w", err)
	}

	return MessageCreated{
		updateWaiter: newUpdateWaiter(),

		Message:    message,
		Literal:    literal,
		MailboxIDs: mailboxIDs,

		Body:      body,
		Structure: structure,
		Envelope:  envelope,
	}, nil
}

func (u MessageCreated) String() string {
	return fmt.Sprintf("MessageCreated: Message.ID = %v, MailboxIDs = %v", utils.ShortID(u.Message.ID), u.MailboxIDs)
}

func (MessageCreated) _isUpdate() {}
