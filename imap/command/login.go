package command

import (
	"fmt"
	rfcparser "github.com/ProtonMail/gluon/rfcparser"
)

type LoginCommand struct {
	UserID   string
	Password string
}

func (l LoginCommand) String() string {
	return fmt.Sprintf("LOGIN '%v' '%v'", l.UserID, l.Password)
}

func (l LoginCommand) SanitizedString() string {
	return fmt.Sprintf("LOGIN '%v' <PASSWORD>", sanitizeString(l.UserID))
}

type LoginCommandParser struct{}

func (LoginCommandParser) FromParser(p *rfcparser.Parser) (Payload, error) {
	// login           = "LOGIN" SP userid SP password
	// userid          = astring
	// password        = astring
	if err := p.Consume(rfcparser.TokenTypeSP, "expected space after command"); err != nil {
		return nil, err
	}

	user, err := p.ParseAString()
	if err != nil {
		return nil, err
	}

	if err := p.Consume(rfcparser.TokenTypeSP, "expected space after userid"); err != nil {
		return nil, err
	}

	password, err := p.ParseAString()
	if err != nil {
		return nil, err
	}

	return &LoginCommand{
		UserID:   user.Value,
		Password: password.Value,
	}, nil
}
