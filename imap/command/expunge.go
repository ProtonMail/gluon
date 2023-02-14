package command

import (
	"fmt"
	"github.com/ProtonMail/gluon/rfcparser"
)

type ExpungeCommand struct{}

func (l ExpungeCommand) String() string {
	return fmt.Sprintf("EXPUNGE")
}

func (l ExpungeCommand) SanitizedString() string {
	return l.String()
}

type ExpungeCommandParser struct{}

func (ExpungeCommandParser) FromParser(p *rfcparser.Parser) (Payload, error) {
	return &ExpungeCommand{}, nil
}
