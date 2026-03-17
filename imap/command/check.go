package command

import (
	"github.com/ProtonMail/gluon/rfcparser"
)

type Check struct{}

func (l Check) String() string {
	return "CHECK"
}

func (l Check) SanitizedString() string {
	return l.String()
}

type CheckCommandParser struct{}

func (CheckCommandParser) FromParser(p *rfcparser.Parser) (Payload, error) {
	return &Check{}, nil
}
