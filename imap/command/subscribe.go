package command

import (
	"fmt"
	"github.com/ProtonMail/gluon/imap/parser"
)

type SubscribeCommand struct {
	Mailbox string
}

func (l SubscribeCommand) String() string {
	return fmt.Sprintf("SUBSCRIBE '%v'", l.Mailbox)
}

func (l SubscribeCommand) SanitizedString() string {
	return fmt.Sprintf("SUBSCRIBE '%v'", sanitizeString(l.Mailbox))
}

type SubscribeCommandParser struct{}

func (SubscribeCommandParser) FromParser(p *parser.Parser) (Payload, error) {
	// subscribe          = "SUBSCRIBE" SP mailbox
	if err := p.Consume(parser.TokenTypeSP, "expected space after command"); err != nil {
		return nil, err
	}

	mailbox, err := p.ParseMailbox()
	if err != nil {
		return nil, err
	}

	return &SubscribeCommand{
		Mailbox: mailbox,
	}, nil
}
