package imap

import (
	"fmt"

	"github.com/ProtonMail/gluon/rfc822"
)

type MessagesCreated struct {
	*updateWaiter

	Messages []*MessageCreated
}

type MessageCreated struct {
	Message    Message
	Literal    []byte
	MailboxIDs []string

	Body      string
	Structure string
	Envelope  string
}

func NewMessagesCreated() *MessagesCreated {
	return &MessagesCreated{
		updateWaiter: newUpdateWaiter(),
	}
}

func (u *MessagesCreated) Add(message Message, literal []byte, mailboxIDs []string) error {
	root, err := rfc822.Parse(literal)
	if err != nil {
		return fmt.Errorf("failed to parse message literal: %w", err)
	}

	body, err := Structure(root, false)
	if err != nil {
		return fmt.Errorf("failed to build message body: %w", err)
	}

	structure, err := Structure(root, true)
	if err != nil {
		return fmt.Errorf("failed to build message body structure: %w", err)
	}

	envelope, err := Envelope(root.ParseHeader())
	if err != nil {
		return fmt.Errorf("failed to build message envelope: %w", err)
	}

	u.Messages = append(u.Messages, &MessageCreated{
		Message:    message,
		Literal:    literal,
		MailboxIDs: mailboxIDs,

		Body:      body,
		Structure: structure,
		Envelope:  envelope,
	})

	return nil
}

func (u *MessagesCreated) String() string {
	return fmt.Sprintf("MessagesCreated (length = %v)", len(u.Messages))
}

func (*MessagesCreated) _isUpdate() {}
