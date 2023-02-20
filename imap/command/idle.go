package command

import (
	"fmt"

	"github.com/ProtonMail/gluon/rfcparser"
)

type Idle struct{}

func (l Idle) String() string {
	return fmt.Sprintf("IDLE")
}

func (l Idle) SanitizedString() string {
	return l.String()
}

type IdleCommandParser struct{}

func (IdleCommandParser) FromParser(p *rfcparser.Parser) (Payload, error) {
	return &Idle{}, nil
}
