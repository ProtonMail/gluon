package command

import (
	"fmt"
	"github.com/ProtonMail/gluon/imap/parser"
)

type UnsubscribeCommand struct {
	Mailbox string
}

func (l UnsubscribeCommand) String() string {
	return fmt.Sprintf("UNSUBSCRIBE '%v'", l.Mailbox)
}

func (l UnsubscribeCommand) SanitizedString() string {
	return fmt.Sprintf("UNSUBSCRIBE '%v'", sanitizeString(l.Mailbox))
}

type UnsubscribeCommandParser struct{}

func (UnsubscribeCommandParser) FromParser(p *parser.Parser) (Payload, error) {
	// unsubscribe          = "UNSUBSCRIBE" SP mailbox
	if err := p.Consume(parser.TokenTypeSP, "expected space after command"); err != nil {
		return nil, err
	}

	mailbox, err := p.ParseMailbox()
	if err != nil {
		return nil, err
	}

	return &UnsubscribeCommand{
		Mailbox: mailbox,
	}, nil
}
