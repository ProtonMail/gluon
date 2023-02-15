package command

import (
	"fmt"
	rfcparser "github.com/ProtonMail/gluon/rfcparser"
)

type MoveCommand struct {
	SeqSet  []SeqRange
	Mailbox string
}

func (l MoveCommand) String() string {
	return fmt.Sprintf("MOVE %v '%v'", l.SeqSet, l.Mailbox)
}

func (l MoveCommand) SanitizedString() string {
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

	return &MoveCommand{
		SeqSet:  seqSet,
		Mailbox: mailbox.Value,
	}, nil
}
