package command

import (
	"fmt"

	"github.com/ProtonMail/gluon/rfcparser"
)

type Create struct {
	Mailbox string
}

func (l Create) String() string {
	return fmt.Sprintf("CREATE '%v'", l.Mailbox)
}

func (l Create) SanitizedString() string {
	return fmt.Sprintf("CREATE '%v'", sanitizeString(l.Mailbox))
}

type CreateCommandParser struct{}

func (CreateCommandParser) FromParser(p *rfcparser.Parser) (Payload, error) {
	// create          = "CREATE" SP mailbox
	if err := p.Consume(rfcparser.TokenTypeSP, "expected space after command"); err != nil {
		return nil, err
	}

	mailbox, err := ParseMailbox(p)
	if err != nil {
		return nil, err
	}

	return &Create{
		Mailbox: mailbox.Value,
	}, nil
}
