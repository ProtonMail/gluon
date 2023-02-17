package rfc5322

import (
	"net/mail"

	"github.com/ProtonMail/gluon/rfcparser"
)

type Parser struct {
	source  *BacktrackingByteScanner
	scanner *rfcparser.Scanner
	parser  *rfcparser.Parser
}

type parserStringType int

const (
	parserStringTypeOther parserStringType = iota
	parserStringTypeUnspaced
	parserStringTypeEncoded
)

type parserString struct {
	String rfcparser.String
	Type   parserStringType
}

func ParseAddress(input string) ([]*mail.Address, error) {
	source := NewBacktrackingByteScanner([]byte(input))
	scanner := rfcparser.NewScannerWithReader(source)
	parser := rfcparser.NewParser(scanner)

	p := Parser{
		source:  source,
		scanner: scanner,
		parser:  parser,
	}

	if err := p.parser.Advance(); err != nil {
		return nil, err
	}

	addr, _, err := parseAddress(&p)

	return addr, err
}

func ParseAddressList(input string) ([]*mail.Address, error) {
	source := NewBacktrackingByteScanner([]byte(input))
	scanner := rfcparser.NewScannerWithReader(source)
	parser := rfcparser.NewParser(scanner)

	p := Parser{
		source:  source,
		scanner: scanner,
		parser:  parser,
	}

	if err := p.parser.Advance(); err != nil {
		return nil, err
	}

	return parseAddressList(&p)
}

type ParserState struct {
	scanner BacktrackingByteScannerScope
	parser  rfcparser.ParserState
}

func (p *Parser) SaveState() ParserState {
	scannerScope := p.source.SaveState()

	return ParserState{
		scanner: scannerScope,
		parser:  p.parser.SaveState(),
	}
}

func (p *Parser) RestoreState(s ParserState) {
	p.source.RestoreState(s.scanner)
	p.parser.RestoreState(s.parser)
}
