package command

import (
	"github.com/ProtonMail/gluon/rfcparser"
)

type StartTLS struct{}

func (l StartTLS) String() string {
	return "STARTTLS"
}

func (l StartTLS) SanitizedString() string {
	return l.String()
}

type StartTLSCommandParser struct{}

func (StartTLSCommandParser) FromParser(p *rfcparser.Parser) (Payload, error) {
	return &StartTLS{}, nil
}
