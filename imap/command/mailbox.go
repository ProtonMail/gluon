package command

import (
	"github.com/ProtonMail/gluon/rfcparser"
	"strings"
)

// ParseMailbox parses a mailbox name as defined in RFC 3501.
func ParseMailbox(p *rfcparser.Parser) (rfcparser.String, error) {
	// mailbox = "INBOX" / astring
	astring, err := p.ParseAString()
	if err != nil {
		return rfcparser.String{}, err
	}

	if strings.EqualFold(astring.Value, "INBOX") {
		astring.Value = "INBOX"
	}

	return astring, nil
}
