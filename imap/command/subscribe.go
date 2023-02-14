package command

import (
	"fmt"
	rfcparser2 "github.com/ProtonMail/gluon/rfcparser"
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

func (SubscribeCommandParser) FromParser(p *rfcparser2.Parser) (Payload, error) {
	// subscribe          = "SUBSCRIBE" SP mailbox
	if err := p.Consume(rfcparser2.TokenTypeSP, "expected space after command"); err != nil {
		return nil, err
	}

	mailbox, err := p.ParseMailbox()
	if err != nil {
		return nil, err
	}

	return &SubscribeCommand{
		Mailbox: mailbox.Value,
	}, nil
}
