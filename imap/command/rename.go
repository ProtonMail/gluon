package command

import (
	"fmt"

	"github.com/ProtonMail/gluon/rfcparser"
)

type Rename struct {
	From string
	To   string
}

func (l Rename) String() string {
	return fmt.Sprintf("RENAME '%v' '%v'", l.From, l.To)
}

func (l Rename) SanitizedString() string {
	return fmt.Sprintf("RENAME '%v' '%v'", sanitizeString(l.From), sanitizeString(l.To))
}

type RenameCommandParser struct{}

func (RenameCommandParser) FromParser(p *rfcparser.Parser) (Payload, error) {
	// rename          = "RENAME" SP mailbox SP mailbox
	if err := p.Consume(rfcparser.TokenTypeSP, "expected space after command"); err != nil {
		return nil, err
	}

	mailboxFrom, err := ParseMailbox(p)
	if err != nil {
		return nil, err
	}

	if err := p.Consume(rfcparser.TokenTypeSP, "expected space after mailbox"); err != nil {
		return nil, err
	}

	mailboxTo, err := ParseMailbox(p)
	if err != nil {
		return nil, err
	}

	return &Rename{
		From: mailboxFrom.Value,
		To:   mailboxTo.Value,
	}, nil
}
