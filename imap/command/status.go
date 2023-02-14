package command

import (
	"fmt"
	"github.com/ProtonMail/gluon/imap/parser"
	"github.com/bradenaw/juniper/xslices"
	"strings"
)

type StatusAttribute int

const (
	StatusAttributeMessages StatusAttribute = iota
	StatusAttributeRecent
	StatusAttributeUIDNext
	StatusAttributeUIDValidity
	StatusAttributeUnseen
)

func (s StatusAttribute) String() string {
	switch s {
	case StatusAttributeRecent:
		return "RECENT"
	case StatusAttributeMessages:
		return "MESSAGES"
	case StatusAttributeUIDNext:
		return "UIDNEXT"
	case StatusAttributeUIDValidity:
		return "UIDVALIDITY"
	case StatusAttributeUnseen:
		return "UNSEEN"
	default:
		return "UNKNOWN"
	}
}

type StatusCommand struct {
	Mailbox    string
	Attributes []StatusAttribute
}

func (s StatusCommand) String() string {
	return fmt.Sprintf("Status '%v' '%v'", s.Mailbox, xslices.Map(s.Attributes, func(s StatusAttribute) string {
		return s.String()
	}))
}

func (s StatusCommand) SanitizedString() string {
	return s.String()
}

type StatusCommandParser struct{}

func (StatusCommandParser) FromParser(p *parser.Parser) (Payload, error) {
	//status          = "STATUS" SP mailbox SP
	//                  "(" status-att *(SP status-att) ")"
	if err := p.Consume(parser.TokenTypeSP, "expected space after command"); err != nil {
		return nil, err
	}

	mailbox, err := p.ParseMailbox()
	if err != nil {
		return nil, err
	}

	if err := p.Consume(parser.TokenTypeSP, "expected space after mailbox"); err != nil {
		return nil, err
	}

	if err := p.Consume(parser.TokenTypeLParen, "expected ( for status attributes start"); err != nil {
		return nil, err
	}

	var attributes []StatusAttribute

	// First attribute.
	{
		attr, err := parseStatusAttribute(p)
		if err != nil {
			return nil, err
		}

		attributes = append(attributes, attr)
	}

	// Remaining.
	for {
		if ok, err := p.Matches(parser.TokenTypeSP); err != nil {
			return nil, err
		} else if !ok {
			break
		}

		attr, err := parseStatusAttribute(p)
		if err != nil {
			return nil, err
		}

		attributes = append(attributes, attr)
	}

	if err := p.Consume(parser.TokenTypeRParen, "expected ) for status attributes end"); err != nil {
		return nil, err
	}

	return &StatusCommand{
		Mailbox:    mailbox,
		Attributes: attributes,
	}, nil
}

func parseStatusAttribute(p *parser.Parser) (StatusAttribute, error) {
	//status-att      = "MESSAGES" / "RECENT" / "UIDNEXT" / "UIDVALIDITY" /
	//                   "UNSEEN"
	attribute, err := p.CollectBytesWhileMatches(parser.TokenTypeChar)
	if err != nil {
		return 0, err
	}

	attributeStr := strings.ToLower(string(attribute))
	switch attributeStr {
	case "messages":
		return StatusAttributeMessages, nil
	case "recent":
		return StatusAttributeRecent, nil
	case "uidnext":
		return StatusAttributeUIDNext, nil
	case "uidvalidity":
		return StatusAttributeUIDValidity, nil
	case "unseen":
		return StatusAttributeUnseen, nil
	default:
		return 0, fmt.Errorf("unknown status attribute '%v'", attributeStr)
	}
}
