package command

import (
	"fmt"
	"github.com/ProtonMail/gluon/rfcparser"
)

type IdleCommand struct{}

func (l IdleCommand) String() string {
	return fmt.Sprintf("IDLE")
}

func (l IdleCommand) SanitizedString() string {
	return l.String()
}

type IdleCommandParser struct{}

func (IdleCommandParser) FromParser(p *rfcparser.Parser) (Payload, error) {
	return &IdleCommand{}, nil
}
