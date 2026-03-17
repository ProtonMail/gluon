package command

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/ProtonMail/gluon/rfcparser"
)

type Authenticate Login

func (l Authenticate) String() string {
	return fmt.Sprintf("AUTHENTICATE '%v' '%v'", l.UserID, l.Password)
}

func (l Authenticate) SanitizedString() string {
	return "AUTHENTICATE <AUTH_DATA>"
}

type AuthenticateCommandParser struct{}

const (
	messageClientAbortedAuthentication        = "client aborted authentication"
	messageInvalidBase64Content               = "invalid base64 content"
	messageUnsupportedAuthenticationMechanism = "unsupported authentication mechanism"
	messageInvalidAuthenticationData          = "invalid authentication data" //nolint:gosec
)

func (AuthenticateCommandParser) FromParser(p *rfcparser.Parser) (Payload, error) {
	// authenticate = "AUTHENTICATE" SP auth-type CRLF base64
	// auth-type    = atom
	// base64       = base64 encoded string
	if err := p.Consume(rfcparser.TokenTypeSP, "expected space after command"); err != nil {
		return nil, err
	}

	method, err := p.ParseAtom()
	if err != nil {
		return nil, err
	}

	if !strings.EqualFold(method, "plain") {
		return nil, p.MakeError(messageUnsupportedAuthenticationMechanism)
	}

	return parseAuthInputString(p)
}

func parseAuthInputString(p *rfcparser.Parser) (*Authenticate, error) {
	// The continued response for the AUTHENTICATE can be whether
	// `*` , indicating the user aborted the authentication
	// a base64 encoded string of the form `identity\0userid\0password`. identity is ignored in IMAP. Some client (Thunderbird) will leave it empty),
	// other will use the userID (Apple Mail).
	parsed, err := p.ParseStringAfterContinuation("")
	if err != nil {
		return nil, err
	}

	input := parsed.Value
	if input == "*" && p.Check(rfcparser.TokenTypeCR) { // behave like dovecot: no extra whitespaces allowed after * when cancelling.
		return nil, p.MakeError(messageClientAbortedAuthentication)
	}

	decoded, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return nil, p.MakeError(messageInvalidBase64Content)
	}

	if len(decoded) < 2 { // min acceptable message be empty username and password (`\x00\x00`).
		return nil, p.MakeError(messageInvalidAuthenticationData)
	}

	split := bytes.Split(decoded[0:], []byte{0})
	if len(split) != 3 {
		return nil, p.MakeError(messageInvalidAuthenticationData)
	}

	return &Authenticate{
		UserID:   string(split[1]),
		Password: string(split[2]),
	}, nil
}
