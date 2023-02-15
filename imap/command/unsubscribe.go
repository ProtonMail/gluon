package command

import (
	"fmt"
	rfcparser "github.com/ProtonMail/gluon/rfcparser"
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

func (UnsubscribeCommandParser) FromParser(p *rfcparser.Parser) (Payload, error) {
	// unsubscribe          = "UNSUBSCRIBE" SP mailbox
	if err := p.Consume(rfcparser.TokenTypeSP, "expected space after command"); err != nil {
		return nil, err
	}

	mailbox, err := ParseMailbox(p)
	if err != nil {
		return nil, err
	}

	return &UnsubscribeCommand{
		Mailbox: mailbox.Value,
	}, nil
}
