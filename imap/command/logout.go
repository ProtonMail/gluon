package command

import (
	"github.com/ProtonMail/gluon/rfcparser"
)

type Logout struct{}

func (l Logout) String() string {
	return "LOGOUT"
}

func (l Logout) SanitizedString() string {
	return l.String()
}

type LogoutCommandParser struct{}

func (LogoutCommandParser) FromParser(p *rfcparser.Parser) (Payload, error) {
	return &Logout{}, nil
}
