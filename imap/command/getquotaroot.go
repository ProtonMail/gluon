package command

import (
	"fmt"

	"github.com/ProtonMail/gluon/rfcparser"
)

type GetQuotaRoot struct {
	Mailbox string
}

func (g GetQuotaRoot) String() string {
	return fmt.Sprintf("GETQUOTAROOT '%v'", g.Mailbox)
}

func (g GetQuotaRoot) SanitizedString() string {
	return g.String()
}

type GetQuotaRootCommandParser struct{}

func (GetQuotaRootCommandParser) FromParser(p *rfcparser.Parser) (Payload, error) {
	// getquotaroot = "GETQUOTAROOT" SP mailbox
	if err := p.Consume(rfcparser.TokenTypeSP, "expected space after command"); err != nil {
		return nil, err
	}

	mailbox, err := ParseMailbox(p)
	if err != nil {
		return nil, err
	}

	return &GetQuotaRoot{
		Mailbox: mailbox.Value,
	}, nil
}
