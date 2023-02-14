package command

import (
	"fmt"
	"github.com/ProtonMail/gluon/rfcparser"
)

type StartTLSCommand struct{}

func (l StartTLSCommand) String() string {
	return fmt.Sprintf("STARTTLS")
}

func (l StartTLSCommand) SanitizedString() string {
	return l.String()
}

type StartTLSCommandParser struct{}

func (StartTLSCommandParser) FromParser(p *rfcparser.Parser) (Payload, error) {
	return &StartTLSCommand{}, nil
}
