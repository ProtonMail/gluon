package command

import (
	"fmt"
	"github.com/ProtonMail/gluon/imap/parser"
)

type IdleCommand struct{}

func (l IdleCommand) String() string {
	return fmt.Sprintf("IDLE")
}

func (l IdleCommand) SanitizedString() string {
	return l.String()
}

type IdleCommandParser struct{}

func (IdleCommandParser) FromParser(p *parser.Parser) (Payload, error) {
	return &IdleCommand{}, nil
}
