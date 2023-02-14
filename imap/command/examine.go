package command

import (
	"fmt"
	"github.com/ProtonMail/gluon/imap/parser"
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

func (ExamineCommandParser) FromParser(p *parser.Parser) (Payload, error) {
	// examine          = "EXAMINE" SP mailbox
	if err := p.Consume(parser.TokenTypeSP, "expected space after command"); err != nil {
		return nil, err
	}

	mailbox, err := p.ParseMailbox()
	if err != nil {
		return nil, err
	}

	return &ExamineCommand{
		Mailbox: mailbox,
	}, nil
}
