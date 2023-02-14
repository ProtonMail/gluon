package command

import (
	"fmt"
	"github.com/ProtonMail/gluon/rfcparser"
)

type UnselectCommand struct{}

func (l UnselectCommand) String() string {
	return fmt.Sprintf("UNSELECT")
}

func (l UnselectCommand) SanitizedString() string {
	return l.String()
}

type UnselectCommandParser struct{}

func (UnselectCommandParser) FromParser(p *rfcparser.Parser) (Payload, error) {
	return &UnselectCommand{}, nil
}
