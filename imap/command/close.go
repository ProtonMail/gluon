package command

import (
	"fmt"
	"github.com/ProtonMail/gluon/imap/parser"
)

type CloseCommand struct{}

func (l CloseCommand) String() string {
	return fmt.Sprintf("CLOSE")
}

func (l CloseCommand) SanitizedString() string {
	return l.String()
}

type CloseCommandParser struct{}

func (CloseCommandParser) FromParser(p *parser.Parser) (Payload, error) {
	return &CloseCommand{}, nil
}
