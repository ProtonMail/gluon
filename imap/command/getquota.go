package command

import (
	"fmt"

	"github.com/ProtonMail/gluon/rfcparser"
)

type GetQuota struct {
	Root string
}

func (g GetQuota) String() string {
	return fmt.Sprintf("GETQUOTA '%v'", g.Root)
}

func (g GetQuota) SanitizedString() string {
	return g.String()
}

type GetQuotaCommandParser struct{}

func (GetQuotaCommandParser) FromParser(p *rfcparser.Parser) (Payload, error) {
	// getquota = "GETQUOTA" SP astring
	if err := p.Consume(rfcparser.TokenTypeSP, "expected space after command"); err != nil {
		return nil, err
	}

	root, err := p.ParseAString()
	if err != nil {
		return nil, err
	}

	return &GetQuota{
		Root: root.Value,
	}, nil
}
