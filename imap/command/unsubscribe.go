package command

import (
	"fmt"

	"github.com/ProtonMail/gluon/rfcparser"
)

type Unsubscribe struct {
	Mailbox string
}

func (l Unsubscribe) String() string {
	return fmt.Sprintf("UNSUBSCRIBE '%v'", l.Mailbox)
}

func (l Unsubscribe) SanitizedString() string {
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

	return &Unsubscribe{
		Mailbox: mailbox.Value,
	}, nil
}
