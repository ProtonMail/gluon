package command

import (
	"fmt"
	"github.com/ProtonMail/gluon/rfcparser"
)

type Copy struct {
	SeqSet  []SeqRange
	Mailbox string
}

func (l Copy) String() string {
	return fmt.Sprintf("COPY %v '%v'", l.SeqSet, l.Mailbox)
}

func (l Copy) SanitizedString() string {
	return fmt.Sprintf("COPY %v '%v'", l.SeqSet, sanitizeString(l.Mailbox))
}

type CopyCommandParser struct{}

func (CopyCommandParser) FromParser(p *rfcparser.Parser) (Payload, error) {
	// copy            = "COPY" SP sequence-set SP mailbox
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

	return &Copy{
		SeqSet:  seqSet,
		Mailbox: mailbox.Value,
	}, nil
}
