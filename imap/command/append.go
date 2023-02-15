package command

import (
	"fmt"
	"github.com/ProtonMail/gluon/rfcparser"
	"time"
)

type Append struct {
	Mailbox  string
	Flags    []string
	DateTime time.Time
	Literal  []byte
}

func (l Append) String() string {
	return fmt.Sprintf("APPEND '%v' Flags='%v' DateTime='%v' Literal=%v",
		l.Mailbox,
		l.Flags,
		l.DateTime,
		l.Literal,
	)
}

func (l Append) SanitizedString() string {
	return fmt.Sprintf("APPEND '%v' Flags='%v' DateTime='%v'",
		sanitizeString(l.Mailbox),
		l.Flags,
		l.DateTime,
	)
}

func (l Append) HasDateTime() bool {
	return l.DateTime != time.Time{}
}

type AppendCommandParser struct{}

func (AppendCommandParser) FromParser(p *rfcparser.Parser) (Payload, error) {
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

	var appendFlags []string

	// check if we have flags.
	flagList, hasFlagList, err := TryParseFlagList(p)
	if err != nil {
		return nil, err
	} else if hasFlagList {
		appendFlags = flagList
	}

	if hasFlagList {
		if err := p.Consume(rfcparser.TokenTypeSP, "expected space after flag list"); err != nil {
			return nil, err
		}
	}

	var dateTime time.Time
	// check date time.
	if !p.Check(rfcparser.TokenTypeLCurly) {
		dt, err := ParseDateTime(p)
		if err != nil {
			return nil, err
		}

		dateTime = dt

		if err := p.Consume(rfcparser.TokenTypeSP, "expected space after flag list"); err != nil {
			return nil, err
		}
	}

	// read literal.
	literal, err := p.ParseLiteral()
	if err != nil {
		return nil, err
	}

	return &Append{
		Mailbox:  mailbox.Value,
		Literal:  literal,
		Flags:    appendFlags,
		DateTime: dateTime,
	}, nil
}
