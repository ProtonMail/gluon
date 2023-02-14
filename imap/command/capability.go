package command

import (
	"fmt"
	"github.com/ProtonMail/gluon/imap/parser"
)

type CapabilityCommand struct{}

func (l CapabilityCommand) String() string {
	return fmt.Sprintf("CAPABILITY")
}

func (l CapabilityCommand) SanitizedString() string {
	return l.String()
}

type CapabilityCommandParser struct{}

func (CapabilityCommandParser) FromParser(p *parser.Parser) (Payload, error) {
	return &CapabilityCommand{}, nil
}
