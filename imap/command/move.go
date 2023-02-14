package command

import (
	"fmt"
	"github.com/ProtonMail/gluon/imap/parser"
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

func (MoveCommandParser) FromParser(p *parser.Parser) (Payload, error) {
	// move            = "MOVE" SP sequence-set SP mailbox
	if err := p.Consume(parser.TokenTypeSP, "expected space after command"); err != nil {
		return nil, err
	}

	seqSet, err := ParseSeqSet(p)
	if err != nil {
		return nil, err
	}

	if err := p.Consume(parser.TokenTypeSP, "expected space after seqset"); err != nil {
		return nil, err
	}

	mailbox, err := p.ParseMailbox()
	if err != nil {
		return nil, err
	}

	return &MoveCommand{
		SeqSet:  seqSet,
		Mailbox: mailbox,
	}, nil
}
