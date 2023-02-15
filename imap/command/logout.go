package command

import (
	"fmt"
	"github.com/ProtonMail/gluon/rfcparser"
)

type Logout struct{}

func (l Logout) String() string {
	return fmt.Sprintf("LOGOUT")
}

func (l Logout) SanitizedString() string {
	return l.String()
}

type LogoutCommandParser struct{}

func (LogoutCommandParser) FromParser(p *rfcparser.Parser) (Payload, error) {
	return &Logout{}, nil
}
