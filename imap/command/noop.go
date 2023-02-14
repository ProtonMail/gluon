package command

import (
	"fmt"
	"github.com/ProtonMail/gluon/imap/parser"
)

type NoopCommand struct{}

func (l NoopCommand) String() string {
	return fmt.Sprintf("NOOP")
}

func (l NoopCommand) SanitizedString() string {
	return l.String()
}

type NoopCommandParser struct{}

func (NoopCommandParser) FromParser(p *parser.Parser) (Payload, error) {
	return &NoopCommand{}, nil
}
