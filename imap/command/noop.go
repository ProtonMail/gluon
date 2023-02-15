package command

import (
	"fmt"
	"github.com/ProtonMail/gluon/rfcparser"
)

type Noop struct{}

func (l Noop) String() string {
	return fmt.Sprintf("NOOP")
}

func (l Noop) SanitizedString() string {
	return l.String()
}

type NoopCommandParser struct{}

func (NoopCommandParser) FromParser(p *rfcparser.Parser) (Payload, error) {
	return &Noop{}, nil
}
