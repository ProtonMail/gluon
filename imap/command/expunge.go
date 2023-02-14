package command

import (
	"fmt"
	"github.com/ProtonMail/gluon/imap/parser"
)

type ExpungeCommand struct{}

func (l ExpungeCommand) String() string {
	return fmt.Sprintf("EXPUNGE")
}

func (l ExpungeCommand) SanitizedString() string {
	return l.String()
}

type ExpungeCommandParser struct{}

func (ExpungeCommandParser) FromParser(p *parser.Parser) (Payload, error) {
	return &ExpungeCommand{}, nil
}
