package command

import (
	"fmt"
	rfcparser2 "github.com/ProtonMail/gluon/rfcparser"
)

type ExamineCommand struct {
	Mailbox string
}

func (l ExamineCommand) String() string {
	return fmt.Sprintf("EXAMINE '%v'", l.Mailbox)
}

func (l ExamineCommand) SanitizedString() string {
	return fmt.Sprintf("EXAMINE '%v'", sanitizeString(l.Mailbox))
}

type ExamineCommandParser struct{}

func (ExamineCommandParser) FromParser(p *rfcparser2.Parser) (Payload, error) {
	// examine          = "EXAMINE" SP mailbox
	if err := p.Consume(rfcparser2.TokenTypeSP, "expected space after command"); err != nil {
		return nil, err
	}

	mailbox, err := ParseMailbox(p)
	if err != nil {
		return nil, err
	}

	return &ExamineCommand{
		Mailbox: mailbox.Value,
	}, nil
}
