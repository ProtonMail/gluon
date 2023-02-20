package command

import (
	"fmt"

	"github.com/ProtonMail/gluon/rfcparser"
)

type Examine struct {
	Mailbox string
}

func (l Examine) String() string {
	return fmt.Sprintf("EXAMINE '%v'", l.Mailbox)
}

func (l Examine) SanitizedString() string {
	return fmt.Sprintf("EXAMINE '%v'", sanitizeString(l.Mailbox))
}

type ExamineCommandParser struct{}

func (ExamineCommandParser) FromParser(p *rfcparser.Parser) (Payload, error) {
	// examine          = "EXAMINE" SP mailbox
	if err := p.Consume(rfcparser.TokenTypeSP, "expected space after command"); err != nil {
		return nil, err
	}

	mailbox, err := ParseMailbox(p)
	if err != nil {
		return nil, err
	}

	return &Examine{
		Mailbox: mailbox.Value,
	}, nil
}
