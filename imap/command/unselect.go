package command

import (
	"fmt"

	"github.com/ProtonMail/gluon/rfcparser"
)

type Unselect struct{}

func (l Unselect) String() string {
	return fmt.Sprintf("UNSELECT")
}

func (l Unselect) SanitizedString() string {
	return l.String()
}

type UnselectCommandParser struct{}

func (UnselectCommandParser) FromParser(p *rfcparser.Parser) (Payload, error) {
	return &Unselect{}, nil
}
