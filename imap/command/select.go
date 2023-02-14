package command

import (
	"fmt"
	"github.com/ProtonMail/gluon/imap/parser"
)

type SelectCommand struct {
	Mailbox string
}

func (l SelectCommand) String() string {
	return fmt.Sprintf("SELECT '%v'", l.Mailbox)
}

func (l SelectCommand) SanitizedString() string {
	return fmt.Sprintf("SELECT '%v'", sanitizeString(l.Mailbox))
}

type SelectCommandParser struct{}

func (SelectCommandParser) FromParser(p *parser.Parser) (Payload, error) {
	// select          = "SELECT" SP mailbox
	if err := p.Consume(parser.TokenTypeSP, "expected space after command"); err != nil {
		return nil, err
	}

	mailbox, err := p.ParseMailbox()
	if err != nil {
		return nil, err
	}

	return &SelectCommand{
		Mailbox: mailbox,
	}, nil
}
