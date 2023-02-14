package command

import (
	"fmt"
	rfcparser2 "github.com/ProtonMail/gluon/rfcparser"
	"strings"
)

type FetchAttribute interface {
	String() string
}

type FetchCommand struct {
	SeqSet     []SeqRange
	Attributes []FetchAttribute
}

func (f FetchCommand) String() string {
	return fmt.Sprintf("FETCH %v %v", f.SeqSet, f.Attributes)
}

func (f FetchCommand) SanitizedString() string {
	return f.String()
}

type FetchCommandParser struct{}

func (FetchCommandParser) FromParser(p *rfcparser2.Parser) (Payload, error) {
	//fetch           = "FETCH" SP sequence-set SP ("ALL" / "FULL" / "FAST" /
	//                  fetch-att / "(" fetch-att *(SP fetch-att) ")")
	if err := p.Consume(rfcparser2.TokenTypeSP, "expected space after command"); err != nil {
		return nil, err
	}

	seqSet, err := ParseSeqSet(p)
	if err != nil {
		return nil, err
	}

	if err := p.Consume(rfcparser2.TokenTypeSP, "expected space after sequence set"); err != nil {
		return nil, err
	}

	var attributes []FetchAttribute

	if p.Check(rfcparser2.TokenTypeLParen) {
		// Multiple list of attributes.
		attr, err := parseFetchAttributes(p)
		if err != nil {
			return nil, err
		}

		attributes = attr
	} else {
		// One single attribute.
		attribute, err := parseFetchAttributeName(p)
		if err != nil {
			return nil, err
		}

		switch attribute {
		case "all":
			attributes = []FetchAttribute{&FetchAttributeAll{}}
		case "full":
			attributes = []FetchAttribute{&FetchAttributeFull{}}
		case "fast":
			attributes = []FetchAttribute{&FetchAttributeFast{}}
		default:
			attr, err := handleFetchAttribute(attribute, p)
			if err != nil {
				return nil, err
			}

			attributes = []FetchAttribute{attr}
		}
	}

	return &FetchCommand{SeqSet: seqSet, Attributes: attributes}, nil
}

func parseFetchAttributeName(p *rfcparser2.Parser) (string, error) {
	att, err := p.CollectBytesWhileMatches(rfcparser2.TokenTypeChar)
	if err != nil {
		return "", err
	}

	return strings.ToLower(string(att.Value)), nil
}

func parseFetchAttributes(p *rfcparser2.Parser) ([]FetchAttribute, error) {
	var attributes []FetchAttribute

	if err := p.Consume(rfcparser2.TokenTypeLParen, "expected ( for fetch attribute list start"); err != nil {
		return nil, err
	}

	// First attribute.
	{
		attribute, err := parseFetchAttribute(p)
		if err != nil {
			return nil, err
		}

		attributes = append(attributes, attribute)
	}

	// Remaining.
	for {
		if ok, err := p.Matches(rfcparser2.TokenTypeSP); err != nil {
			return nil, err
		} else if !ok {
			break
		}

		attribute, err := parseFetchAttribute(p)
		if err != nil {
			return nil, err
		}

		attributes = append(attributes, attribute)
	}

	if err := p.Consume(rfcparser2.TokenTypeRParen, "expected ) for fetch attribute list end"); err != nil {
		return nil, err
	}

	return attributes, nil
}

func parseFetchAttribute(p *rfcparser2.Parser) (FetchAttribute, error) {
	attribute, err := parseFetchAttributeName(p)
	if err != nil {
		return nil, err
	}

	attr, err := handleFetchAttribute(attribute, p)
	if err != nil {
		return nil, err
	}

	return attr, nil
}

func handleFetchAttribute(name string, p *rfcparser2.Parser) (FetchAttribute, error) {
	/*
	 fetch-att       = "ENVELOPE" / "FLAGS" / "INTERNALDATE" /
	                    "RFC822" [".HEADER" / ".SIZE" / ".TEXT"] /
	                    "BODY" ["STRUCTURE"] / "UID" /
	                    "BODY" section ["<" number "." nz-number ">"] /
	                    "BODY.PEEK" section ["<" number "." nz-number ">"]
	*/
	switch name {
	case "envelope":
		return &FetchAttributeEnvelope{}, nil
	case "flags":
		return &FetchAttributeFlags{}, nil
	case "internaldate":
		return &FetchAttributeInternalDate{}, nil
	case "bodystructure":
		return &FetchAttributeBodyStructure{}, nil
	case "uid":
		return &FetchAttributeUID{}, nil
	case "rfc":
		return handleRFC822FetchAttribute(p)
	case "body":
		return handleBodyFetchAttribute(p)
	default:
		return nil, fmt.Errorf("unknown fetch attribute '%v'", name)
	}
}

func handleRFC822FetchAttribute(p *rfcparser2.Parser) (FetchAttribute, error) {
	if err := p.ConsumeBytesFold('8', '2', '2'); err != nil {
		return nil, err
	}

	if err := p.Consume(rfcparser2.TokenTypePeriod, "expected '.' after RFC822 fetch attribute"); err != nil {
		return nil, err
	}

	attribute, err := parseFetchAttributeName(p)
	if err != nil {
		return nil, err
	}

	switch attribute {
	case "header":
		return &FetchAttributeRFC822Header{}, nil
	case "size":
		return &FetchAttributeRFC822Size{}, nil
	case "text":
		return &FetchAttributeRFC822Text{}, nil
	default:
		return nil, fmt.Errorf("unknown fetch attribute 'RFC822.%v", attribute)
	}
}

func handleBodyFetchAttribute(p *rfcparser2.Parser) (FetchAttribute, error) {
	// Check if we are dealing with BODY only
	if !p.Check(rfcparser2.TokenTypeLBracket) && !p.Check(rfcparser2.TokenTypePeriod) {
		return &FetchAttributeBody{}, nil
	}

	var readOnly = false

	if ok, err := p.Matches(rfcparser2.TokenTypePeriod); err != nil {
		return nil, err
	} else if ok {
		if err := p.ConsumeBytesFold('P', 'E', 'E', 'K'); err != nil {
			return nil, err
		}

		readOnly = true
	}

	if err := p.Consume(rfcparser2.TokenTypeLBracket, "expected [ for body section start"); err != nil {
		return nil, err
	}

	var section BodySection

	if !p.Check(rfcparser2.TokenTypeRBracket) {
		s, err := parseSectionSpec(p)
		if err != nil {
			return nil, err
		}

		section = s
	}

	if err := p.Consume(rfcparser2.TokenTypeRBracket, "expected ] for body section end"); err != nil {
		return nil, err
	}

	var partial *BodySectionPartial

	if ok, err := p.Matches(rfcparser2.TokenTypeLess); err != nil {
		return nil, err
	} else if ok {
		offset, err := p.ParseNumber()
		if err != nil {
			return nil, err
		}

		if err := p.Consume(rfcparser2.TokenTypePeriod, "expected '.' after partial start"); err != nil {
			return nil, err
		}

		count, err := ParseNZNumber(p)
		if err != nil {
			return nil, err
		}

		if err := p.Consume(rfcparser2.TokenTypeGreater, "expected > for end of partial specification"); err != nil {
			return nil, err
		}

		partial = &BodySectionPartial{
			Offset: int64(offset),
			Count:  int64(count),
		}
	}

	return &FetchAttributeBodySection{Peek: readOnly, Section: section, Partial: partial}, nil
}

func parseSectionSpec(p *rfcparser2.Parser) (BodySection, error) {
	// section-spec    = section-msgtext / (section-part ["." section-text])
	if p.Check(rfcparser2.TokenTypeDigit) {
		part, err := parseSectionPart(p)
		if err != nil {
			return nil, err
		}

		// Note: trailing . is consumed by parserSectionPart().
		text, err := parseSectionText(p)
		if err != nil {
			return nil, err
		}

		return &BodySectionPart{Part: part, Section: text}, nil
	}

	return parseSectionMsgText(p)
}

func parseSectionPart(p *rfcparser2.Parser) ([]int, error) {
	// section-part    = nz-number *("." nz-number)
	//                     ; body part nesting
	var result []int

	{
		num, err := ParseNZNumber(p)
		if err != nil {
			return nil, err
		}

		result = append(result, num)
	}

	for {
		if ok, err := p.Matches(rfcparser2.TokenTypePeriod); err != nil {
			return nil, err
		} else if !ok {
			break
		}

		// If there is section-text after the section-part we can continue with number processing.
		if !p.Check(rfcparser2.TokenTypeDigit) {
			break
		}

		num, err := ParseNZNumber(p)
		if err != nil {
			return nil, err
		}

		result = append(result, num)
	}

	return result, nil
}

func parseSectionText(p *rfcparser2.Parser) (BodySection, error) {
	// section-text    = section-msgtext / "MIME"
	//                    ; text other than actual body part (headers, etc.)
	text, err := p.CollectBytesWhileMatches(rfcparser2.TokenTypeChar)
	if err != nil {
		return nil, err
	}

	textStr := strings.ToLower(string(text.Value))

	if textStr == "mime" {
		return &BodySectionMIME{}, nil
	}

	return handleSectionMessageText(textStr, p)
}

func parseSectionMsgText(p *rfcparser2.Parser) (BodySection, error) {
	text, err := p.CollectBytesWhileMatches(rfcparser2.TokenTypeChar)
	if err != nil {
		return nil, err
	}

	textStr := strings.ToLower(string(text.Value))

	return handleSectionMessageText(textStr, p)
}

func handleSectionMessageText(text string, p *rfcparser2.Parser) (BodySection, error) {
	// section-msgtext = "HEADER" / "HEADER.FIELDS" [".NOT"] SP header-list /
	//                   "TEXT"
	//                    ; top-level or MESSAGE/RFC822 part
	switch text {
	case "header":
		if ok, err := p.Matches(rfcparser2.TokenTypePeriod); err != nil {
			return nil, err
		} else if !ok {
			return &BodySectionHeader{}, nil
		}

		return parseHeaderFieldsSectionMessageText(p)

	case "text":
		return &BodySectionText{}, nil
	default:
		return nil, fmt.Errorf("unknown section msg text value '%v'", text)
	}
}

func parseHeaderFieldsSectionMessageText(p *rfcparser2.Parser) (BodySection, error) {
	// Read fields bit
	{
		text, err := collectBodySectionText(p)
		if err != nil {
			return nil, err
		}

		if text != "fields" {
			return nil, p.MakeError("expected 'FIELDS' after 'HEADER.'")
		}
	}

	var negate bool

	if ok, err := p.Matches(rfcparser2.TokenTypePeriod); err != nil {
		return nil, err
	} else if ok {
		text, err := collectBodySectionText(p)
		if err != nil {
			return nil, err
		}

		if text != "not" {
			return nil, p.MakeError("expected 'NOT'")
		}

		negate = true
	}

	if err := p.Consume(rfcparser2.TokenTypeSP, "expected space"); err != nil {
		return nil, err
	}

	headerList, err := parseHeaderList(p)
	if err != nil {
		return nil, err
	}

	return &BodySectionHeaderFields{
		Negate: negate,
		Fields: headerList,
	}, nil
}

func collectBodySectionText(p *rfcparser2.Parser) (string, error) {
	text, err := p.CollectBytesWhileMatches(rfcparser2.TokenTypeChar)
	if err != nil {
		return "", err
	}

	return strings.ToLower(string(text.Value)), nil
}

func parseHeaderList(p *rfcparser2.Parser) ([]string, error) {
	var result []string

	if err := p.Consume(rfcparser2.TokenTypeLParen, "expected ( for header list start"); err != nil {
		return nil, err
	}

	{
		header, err := p.ParseAString()
		if err != nil {
			return nil, err
		}

		result = append(result, header.Value)
	}

	for {
		if ok, err := p.Matches(rfcparser2.TokenTypeSP); err != nil {
			return nil, err
		} else if !ok {
			break
		}

		header, err := p.ParseAString()
		if err != nil {
			return nil, err
		}

		result = append(result, header.Value)
	}

	if err := p.Consume(rfcparser2.TokenTypeRParen, "expected ) for header list end"); err != nil {
		return nil, err
	}

	return result, nil
}