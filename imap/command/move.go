package command

import (
	"fmt"

	"github.com/ProtonMail/gluon/rfcparser"
)

type Move struct {
	SeqSet  []SeqRange
	Mailbox string
}

func (l Move) String() string {
	return fmt.Sprintf("MOVE %v '%v'", l.SeqSet, l.Mailbox)
}

func (l Move) SanitizedString() string {
	return fmt.Sprintf("MOVE %v '%v'", l.SeqSet, sanitizeString(l.Mailbox))
}

type MoveCommandParser struct{}

func (MoveCommandParser) FromParser(p *rfcparser.Parser) (Payload, error) {
	// move            = "MOVE" SP sequence-set SP mailbox
	if err := p.Consume(rfcparser.TokenTypeSP, "expected space after command"); err != nil {
		return nil, err
	}

	seqSet, err := ParseSeqSet(p)
	if err != nil {
		return nil, err
	}

	if err := p.Consume(rfcparser.TokenTypeSP, "expected space after seqset"); err != nil {
		return nil, err
	}

	mailbox, err := ParseMailbox(p)
	if err != nil {
		return nil, err
	}

	return &Move{
		SeqSet:  seqSet,
		Mailbox: mailbox.Value,
	}, nil
}
