package command

import (
	"github.com/ProtonMail/gluon/rfcparser"
)

type Close struct{}

func (l Close) String() string {
	return "CLOSE"
}

func (l Close) SanitizedString() string {
	return l.String()
}

type CloseCommandParser struct{}

func (CloseCommandParser) FromParser(p *rfcparser.Parser) (Payload, error) {
	return &Close{}, nil
}
