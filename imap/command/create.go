package command

import (
	"fmt"
	rfcparser2 "github.com/ProtonMail/gluon/rfcparser"
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

func (CreateCommandParser) FromParser(p *rfcparser2.Parser) (Payload, error) {
	// create          = "CREATE" SP mailbox
	if err := p.Consume(rfcparser2.TokenTypeSP, "expected space after command"); err != nil {
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
