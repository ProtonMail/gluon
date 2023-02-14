package command

import (
	"fmt"
	"github.com/ProtonMail/gluon/imap/parser"
)

type RenameCommand struct {
	From string
	To   string
}

func (l RenameCommand) String() string {
	return fmt.Sprintf("RENAME '%v' '%v'", l.From, l.To)
}

func (l RenameCommand) SanitizedString() string {
	return fmt.Sprintf("RENAME '%v' '%v'", sanitizeString(l.From), sanitizeString(l.To))
}

type RenameCommandParser struct{}

func (RenameCommandParser) FromParser(p *parser.Parser) (Payload, error) {
	// rename          = "RENAME" SP mailbox SP mailbox
	if err := p.Consume(parser.TokenTypeSP, "expected space after command"); err != nil {
		return nil, err
	}

	mailboxFrom, err := p.ParseMailbox()
	if err != nil {
		return nil, err
	}

	if err := p.Consume(parser.TokenTypeSP, "expected space after mailbox"); err != nil {
		return nil, err
	}

	mailboxTo, err := p.ParseMailbox()
	if err != nil {
		return nil, err
	}

	return &RenameCommand{
		From: mailboxFrom,
		To:   mailboxTo,
	}, nil
}
