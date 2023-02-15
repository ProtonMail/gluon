package command

import (
	"fmt"
	"github.com/ProtonMail/gluon/rfcparser"
)

type Expunge struct{}

func (l Expunge) String() string {
	return fmt.Sprintf("EXPUNGE")
}

func (l Expunge) SanitizedString() string {
	return l.String()
}

type ExpungeCommandParser struct{}

func (ExpungeCommandParser) FromParser(p *rfcparser.Parser) (Payload, error) {
	return &Expunge{}, nil
}
