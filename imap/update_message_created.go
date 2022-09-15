package imap

import (
	"fmt"

	"github.com/ProtonMail/gluon/rfc822"
)

type ParsedMessage struct {
	Body      string
	Structure string
	Envelope  string
}

func NewParsedMessage(literal []byte) (*ParsedMessage, error) {
	root := rfc822.Parse(literal)

	body, structure, err := Structure(root)
	if err != nil {
		return nil, fmt.Errorf("failed to build message body and structure: %w", err)
	}

	header, err := root.ParseHeader()
	if err != nil {
		return nil, fmt.Errorf("failed to parser message header: %w", err)
	}

	envelope, err := Envelope(header)
	if err != nil {
		return nil, fmt.Errorf("failed to build message envelope: %w", err)
	}

	return &ParsedMessage{
		Body:      body,
		Structure: structure,
		Envelope:  envelope,
	}, nil
}

type MessagesCreated struct {
	updateBase

	*updateWaiter

	Messages []*MessageCreated
}

type MessageCreated struct {
	Message    Message
	Literal    []byte
	MailboxIDs []LabelID

	ParsedMessage *ParsedMessage
}

func NewMessagesCreated() *MessagesCreated {
	return &MessagesCreated{
		updateWaiter: newUpdateWaiter(),
	}
}

func (u *MessagesCreated) Add(message Message, literal []byte, parsedMessage *ParsedMessage, mailboxIDs ...LabelID) {
	u.Messages = append(u.Messages, &MessageCreated{
		Message:       message,
		Literal:       literal,
		MailboxIDs:    mailboxIDs,
		ParsedMessage: parsedMessage,
	})
}

func (u *MessagesCreated) String() string {
	return fmt.Sprintf("MessagesCreated (length = %v)", len(u.Messages))
}
