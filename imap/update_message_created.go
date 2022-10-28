package imap

import (
	"fmt"
	"github.com/ProtonMail/gluon/rfc822"
	"github.com/bradenaw/juniper/xslices"
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
	Message       Message
	Literal       []byte
	MailboxIDs    []MailboxID
	ParsedMessage *ParsedMessage
}

func NewMessagesCreated(updates ...*MessageCreated) *MessagesCreated {
	return &MessagesCreated{
		updateWaiter: newUpdateWaiter(),
		Messages:     updates,
	}
}

func (u *MessagesCreated) String() string {
	return fmt.Sprintf("MessagesCreated: MessageCount=%v Messages=%v",
		len(u.Messages),
		xslices.Map(u.Messages, func(m *MessageCreated) string {
			return fmt.Sprintf("ID:%v Mailboxes:%v Flags:%s",
				m.Message.ID.ShortID(),
				xslices.Map(m.MailboxIDs, func(mboxID MailboxID) string {
					return mboxID.ShortID()
				}),
				m.Message.Flags.ToSlice(),
			)
		}),
	)
}
