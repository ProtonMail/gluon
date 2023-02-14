package command

import (
	"fmt"
	"github.com/ProtonMail/gluon/imap/parser"
)

type LSubCommand struct {
	Mailbox     string
	LSubMailbox string
}

func (l LSubCommand) String() string {
	return fmt.Sprintf("LSUB '%v' '%v'", l.Mailbox, l.LSubMailbox)
}

func (l LSubCommand) SanitizedString() string {
	return l.String()
}

type LSubCommandParser struct{}

func (LSubCommandParser) FromParser(p *parser.Parser) (Payload, error) {
	// lsub            = "LSUB" SP mailbox SP list-mailbox
	if err := p.Consume(parser.TokenTypeSP, "expected space after command"); err != nil {
		return nil, err
	}

	mailbox, err := p.ParseMailbox()
	if err != nil {
		return nil, err
	}

	if err := p.Consume(parser.TokenTypeSP, "expected space after mailbox"); err != nil {
		return nil, err
	}

	listMailbox, err := parseListMailbox(p)
	if err != nil {
		return nil, err
	}

	return &LSubCommand{
		Mailbox:     mailbox,
		LSubMailbox: listMailbox,
	}, nil
}
