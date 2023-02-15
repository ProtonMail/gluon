package command

import (
	"fmt"
	rfcparser "github.com/ProtonMail/gluon/rfcparser"
)

type Delete struct {
	Mailbox string
}

func (l Delete) String() string {
	return fmt.Sprintf("DELETE '%v'", l.Mailbox)
}

func (l Delete) SanitizedString() string {
	return fmt.Sprintf("DELETE '%v'", sanitizeString(l.Mailbox))
}

type DeleteCommandParser struct{}

func (DeleteCommandParser) FromParser(p *rfcparser.Parser) (Payload, error) {
	// delete          = "DELETE" SP mailbox
	if err := p.Consume(rfcparser.TokenTypeSP, "expected space after command"); err != nil {
		return nil, err
	}

	mailbox, err := ParseMailbox(p)
	if err != nil {
		return nil, err
	}

	return &Delete{
		Mailbox: mailbox.Value,
	}, nil
}
