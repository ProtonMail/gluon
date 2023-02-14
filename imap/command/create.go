package command

import (
	"fmt"
	"github.com/ProtonMail/gluon/imap/parser"
)

type CreateCommand struct {
	Mailbox string
}

func (l CreateCommand) String() string {
	return fmt.Sprintf("CREATE '%v'", l.Mailbox)
}

func (l CreateCommand) SanitizedString() string {
	return fmt.Sprintf("CREATE '%v'", sanitizeString(l.Mailbox))
}

type CreateCommandParser struct{}

func (CreateCommandParser) FromParser(p *parser.Parser) (Payload, error) {
	// create          = "CREATE" SP mailbox
	if err := p.Consume(parser.TokenTypeSP, "expected space after command"); err != nil {
		return nil, err
	}

	mailbox, err := p.ParseMailbox()
	if err != nil {
		return nil, err
	}

	return &CreateCommand{
		Mailbox: mailbox,
	}, nil
}
