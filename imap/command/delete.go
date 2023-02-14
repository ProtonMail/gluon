package command

import (
	"fmt"
	rfcparser2 "github.com/ProtonMail/gluon/rfcparser"
)

type DeleteCommand struct {
	Mailbox string
}

func (l DeleteCommand) String() string {
	return fmt.Sprintf("DELETE '%v'", l.Mailbox)
}

func (l DeleteCommand) SanitizedString() string {
	return fmt.Sprintf("DELETE '%v'", sanitizeString(l.Mailbox))
}

type DeleteCommandParser struct{}

func (DeleteCommandParser) FromParser(p *rfcparser2.Parser) (Payload, error) {
	// delete          = "DELETE" SP mailbox
	if err := p.Consume(rfcparser2.TokenTypeSP, "expected space after command"); err != nil {
		return nil, err
	}

	mailbox, err := ParseMailbox(p)
	if err != nil {
		return nil, err
	}

	return &DeleteCommand{
		Mailbox: mailbox.Value,
	}, nil
}
