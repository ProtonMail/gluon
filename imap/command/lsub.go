package command

import (
	"fmt"

	"github.com/ProtonMail/gluon/rfcparser"
)

type LSub struct {
	Mailbox     string
	LSubMailbox string
}

func (l LSub) String() string {
	return fmt.Sprintf("LSUB '%v' '%v'", l.Mailbox, l.LSubMailbox)
}

func (l LSub) SanitizedString() string {
	return l.String()
}

type LSubCommandParser struct{}

func (LSubCommandParser) FromParser(p *rfcparser.Parser) (Payload, error) {
	// lsub            = "LSUB" SP mailbox SP list-mailbox
	if err := p.Consume(rfcparser.TokenTypeSP, "expected space after command"); err != nil {
		return nil, err
	}

	mailbox, err := ParseMailbox(p)
	if err != nil {
		return nil, err
	}

	if err := p.Consume(rfcparser.TokenTypeSP, "expected space after mailbox"); err != nil {
		return nil, err
	}

	listMailbox, err := parseListMailbox(p)
	if err != nil {
		return nil, err
	}

	return &LSub{
		Mailbox:     mailbox.Value,
		LSubMailbox: listMailbox.Value,
	}, nil
}
