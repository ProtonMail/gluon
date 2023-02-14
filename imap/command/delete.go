package command

import (
	"fmt"
	"github.com/ProtonMail/gluon/imap/parser"
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

func (DeleteCommandParser) FromParser(p *parser.Parser) (Payload, error) {
	// delete          = "DELETE" SP mailbox
	if err := p.Consume(parser.TokenTypeSP, "expected space after command"); err != nil {
		return nil, err
	}

	mailbox, err := p.ParseMailbox()
	if err != nil {
		return nil, err
	}

	return &DeleteCommand{
		Mailbox: mailbox,
	}, nil
}
