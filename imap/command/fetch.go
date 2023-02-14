package command

import (
	"fmt"
	"github.com/ProtonMail/gluon/imap/parser"
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

func (FetchCommandParser) FromParser(p *parser.Parser) (Payload, error) {
	//fetch           = "FETCH" SP sequence-set SP ("ALL" / "FULL" / "FAST" /
	//                  fetch-att / "(" fetch-att *(SP fetch-att) ")")
	if err := p.Consume(parser.TokenTypeSP, "expected space after command"); err != nil {
		return nil, err
	}

	seqSet, err := ParseSeqSet(p)
	if err != nil {
		return nil, err
	}

	if err := p.Consume(parser.TokenTypeSP, "expected space after sequence set"); err != nil {
		return nil, err
	}

	var attributes []FetchAttribute

	if p.Check(parser.TokenTypeLParen) {
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

func parseFetchAttributeName(p *parser.Parser) (string, error) {
	att, err := p.CollectBytesWhileMatches(parser.TokenTypeChar)
	if err != nil {
		return "", err
	}

	return strings.ToLower(string(att)), nil
}

func parseFetchAttributes(p *parser.Parser) ([]FetchAttribute, error) {
	var attributes []FetchAttribute

	if err := p.Consume(parser.TokenTypeLParen, "expected ( for fetch attribute list start"); err != nil {
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
		if ok, err := p.Matches(parser.TokenTypeSP); err != nil {
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

	if err := p.Consume(parser.TokenTypeRParen, "expected ) for fetch attribute list end"); err != nil {
		return nil, err
	}

	return attributes, nil
}

func parseFetchAttribute(p *parser.Parser) (FetchAttribute, error) {
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

func handleFetchAttribute(name string, p *parser.Parser) (FetchAttribute, error) {
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

func handleRFC822FetchAttribute(p *parser.Parser) (FetchAttribute, error) {
	if err := p.ConsumeBytesFold('8', '2', '2'); err != nil {
		return nil, err
	}

	if err := p.Consume(parser.TokenTypePeriod, "expected '.' after RFC822 fetch attribute"); err != nil {
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

func handleBodyFetchAttribute(p *parser.Parser) (FetchAttribute, error) {
	// Check if we are dealing with BODY only
	if !p.Check(parser.TokenTypeLBracket) && !p.Check(parser.TokenTypePeriod) {
		return &FetchAttributeBody{}, nil
	}

	var readOnly = false

	if ok, err := p.Matches(parser.TokenTypePeriod); err != nil {
		return nil, err
	} else if ok {
		if err := p.ConsumeBytesFold('P', 'E', 'E', 'K'); err != nil {
			return nil, err
		}

		readOnly = true
	}

	if err := p.Consume(parser.TokenTypeLBracket, "expected [ for body section start"); err != nil {
		return nil, err
	}

	var section BodySection

	if !p.Check(parser.TokenTypeRBracket) {
		s, err := parseSectionSpec(p)
		if err != nil {
			return nil, err
		}

		section = s
	}

	if err := p.Consume(parser.TokenTypeRBracket, "expected ] for body section end"); err != nil {
		return nil, err
	}

	var partial *BodySectionPartial

	if ok, err := p.Matches(parser.TokenTypeLess); err != nil {
		return nil, err
	} else if ok {
		offset, err := p.ParseNumber()
		if err != nil {
			return nil, err
		}

		if err := p.Consume(parser.TokenTypePeriod, "expected '.' after partial start"); err != nil {
			return nil, err
		}

		count, err := ParseNZNumber(p)
		if err != nil {
			return nil, err
		}

		if err := p.Consume(parser.TokenTypeGreater, "expected > for end of partial specification"); err != nil {
			return nil, err
		}

		partial = &BodySectionPartial{
			Offset: int64(offset),
			Count:  int64(count),
		}
	}

	return &FetchAttributeBodySection{Peek: readOnly, Section: section, Partial: partial}, nil
}

func parseSectionSpec(p *parser.Parser) (BodySection, error) {
	// section-spec    = section-msgtext / (section-part ["." section-text])
	if p.Check(parser.TokenTypeDigit) {
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

func parseSectionPart(p *parser.Parser) ([]int, error) {
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
		if ok, err := p.Matches(parser.TokenTypePeriod); err != nil {
			return nil, err
		} else if !ok {
			break
		}

		// If there is section-text after the section-part we can continue with number processing.
		if !p.Check(parser.TokenTypeDigit) {
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

func parseSectionText(p *parser.Parser) (BodySection, error) {
	// section-text    = section-msgtext / "MIME"
	//                    ; text other than actual body part (headers, etc.)
	text, err := p.CollectBytesWhileMatches(parser.TokenTypeChar)
	if err != nil {
		return nil, err
	}

	textStr := strings.ToLower(string(text))

	if textStr == "mime" {
		return &BodySectionMIME{}, nil
	}

	return handleSectionMessageText(textStr, p)
}

func parseSectionMsgText(p *parser.Parser) (BodySection, error) {
	text, err := p.CollectBytesWhileMatches(parser.TokenTypeChar)
	if err != nil {
		return nil, err
	}

	textStr := strings.ToLower(string(text))

	return handleSectionMessageText(textStr, p)
}

func handleSectionMessageText(text string, p *parser.Parser) (BodySection, error) {
	// section-msgtext = "HEADER" / "HEADER.FIELDS" [".NOT"] SP header-list /
	//                   "TEXT"
	//                    ; top-level or MESSAGE/RFC822 part
	switch text {
	case "header":
		if ok, err := p.Matches(parser.TokenTypePeriod); err != nil {
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

func parseHeaderFieldsSectionMessageText(p *parser.Parser) (BodySection, error) {
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

	if ok, err := p.Matches(parser.TokenTypePeriod); err != nil {
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

	if err := p.Consume(parser.TokenTypeSP, "expected space"); err != nil {
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

func collectBodySectionText(p *parser.Parser) (string, error) {
	text, err := p.CollectBytesWhileMatches(parser.TokenTypeChar)
	if err != nil {
		return "", err
	}

	return strings.ToLower(string(text)), nil
}

func parseHeaderList(p *parser.Parser) ([]string, error) {
	var result []string

	if err := p.Consume(parser.TokenTypeLParen, "expected ( for header list start"); err != nil {
		return nil, err
	}

	{
		header, err := p.ParseAString()
		if err != nil {
			return nil, err
		}

		result = append(result, header)
	}

	for {
		if ok, err := p.Matches(parser.TokenTypeSP); err != nil {
			return nil, err
		} else if !ok {
			break
		}

		header, err := p.ParseAString()
		if err != nil {
			return nil, err
		}

		result = append(result, header)
	}

	if err := p.Consume(parser.TokenTypeRParen, "expected ) for header list end"); err != nil {
		return nil, err
	}

	return result, nil
}
