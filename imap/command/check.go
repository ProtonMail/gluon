package command

import (
	"fmt"
	"github.com/ProtonMail/gluon/imap/parser"
)

type CheckCommand struct{}

func (l CheckCommand) String() string {
	return fmt.Sprintf("CHECK")
}

func (l CheckCommand) SanitizedString() string {
	return l.String()
}

type CheckCommandParser struct{}

func (CheckCommandParser) FromParser(p *parser.Parser) (Payload, error) {
	return &CheckCommand{}, nil
}
