package command

import (
	"github.com/ProtonMail/gluon/rfcparser"
)

type Idle struct{}

func (l Idle) String() string {
	return "IDLE"
}

func (l Idle) SanitizedString() string {
	return l.String()
}

type IdleCommandParser struct{}

func (IdleCommandParser) FromParser(p *rfcparser.Parser) (Payload, error) {
	return &Idle{}, nil
}
