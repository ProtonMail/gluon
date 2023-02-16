package command

import (
	"fmt"

	"github.com/ProtonMail/gluon/rfcparser"
	"github.com/bradenaw/juniper/xslices"
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

type Status struct {
	Mailbox    string
	Attributes []StatusAttribute
}

func (s Status) String() string {
	return fmt.Sprintf("Status '%v' '%v'", s.Mailbox, xslices.Map(s.Attributes, func(s StatusAttribute) string {
		return s.String()
	}))
}

func (s Status) SanitizedString() string {
	return s.String()
}

type StatusCommandParser struct{}

func (StatusCommandParser) FromParser(p *rfcparser.Parser) (Payload, error) {
	//status          = "STATUS" SP mailbox SP
	//                  "(" status-att *(SP status-att) ")"
	if err := p.Consume(rfcparser.TokenTypeSP, "expected space after command"); err != nil {
		return nil, err
	}

	mailbox, err := ParseMailbox(p)
	if err != nil {
		return nil, err
	}

	if err := p.Consume(rfcparser.TokenTypeSP, "expected space after mailbox"); err != nil {
		return nil, err
	}

	if err := p.Consume(rfcparser.TokenTypeLParen, "expected ( for status attributes start"); err != nil {
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
		if ok, err := p.Matches(rfcparser.TokenTypeSP); err != nil {
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

	if err := p.Consume(rfcparser.TokenTypeRParen, "expected ) for status attributes end"); err != nil {
		return nil, err
	}

	return &Status{
		Mailbox:    mailbox.Value,
		Attributes: attributes,
	}, nil
}

func parseStatusAttribute(p *rfcparser.Parser) (StatusAttribute, error) {
	//status-att      = "MESSAGES" / "RECENT" / "UIDNEXT" / "UIDVALIDITY" /
	//                   "UNSEEN"
	attribute, err := p.CollectBytesWhileMatches(rfcparser.TokenTypeChar)
	if err != nil {
		return 0, err
	}

	attributeStr := attribute.IntoString().ToLower()
	switch attributeStr.Value {
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
		return 0, p.MakeErrorAtOffset(fmt.Sprintf("unknown status attribute '%v'", attributeStr), attributeStr.Offset)
	}
}
