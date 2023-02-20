package rfcparser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

type TokenType int

const (
	TokenTypeEOF TokenType = iota
	TokenTypeError
	TokenTypeSP
	TokenTypeExclamation
	TokenTypeDQuote
	TokenTypeHash
	TokenTypeDollar
	TokenTypePercent
	TokenTypeAmpersand
	TokenTypeSQuote
	TokenTypeLParen
	TokenTypeRParen
	TokenTypeAsterisk
	TokenTypePlus
	TokenTypeComma
	TokenTypeMinus
	TokenTypePeriod
	TokenTypeSlash
	TokenTypeSemicolon
	TokenTypeColon
	TokenTypeLess
	TokenTypeEqual
	TokenTypeGreater
	TokenTypeQuestion
	TokenTypeAt
	TokenTypeLBracket
	TokenTypeRBracket
	TokenTypeCaret
	TokenTypeUnderscore
	TokenTyeBacktick
	TokenTypeLCurly
	TokenTypePipe
	TokenTypeRCurly
	TokenTypeTilde
	TokenTypeBackslash
	TokenTypeDigit
	TokenTypeChar
	TokenTypeExtendedChar
	TokenTypeCR
	TokenTypeLF
	TokenTypeCTL
)

type Token struct {
	TType  TokenType
	Value  byte
	Offset int
}

type Scanner struct {
	source      Reader
	currentByte byte
	offset      int
}

type Reader interface {
	io.Reader
	ReadByte() (byte, error)
	ReadBytes(byte) ([]byte, error)
}

func NewScanner(source io.Reader) *Scanner {
	return &Scanner{
		source: bufio.NewReader(source),
	}
}

func NewScannerWithReader(source Reader) *Scanner {
	return &Scanner{
		source: source,
	}
}

func (s *Scanner) ConsumeBytes(dst []byte) error {
	// We have already read a byte at this point, so we need to
	// skip this one.
	dst[0] = s.currentByte

	if _, err := io.ReadFull(s.source, dst[1:]); err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) {
			return io.EOF
		}

		return err
	}

	s.offset += len(dst) - 1

	return nil
}

func (s *Scanner) ConsumeUntilNewLine() ([]byte, error) {
	return s.source.ReadBytes('\n')
}

func (s *Scanner) ScanToken() (Token, error) {
	b, err := s.advance()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return s.makeEOF(), nil
		}

		return Token{}, nil
	}

	if isByteDigit(b) {
		return s.makeToken(TokenTypeDigit), nil
	}

	if isByteAlpha(b) {
		return s.makeToken(TokenTypeChar), nil
	}

	if isByteExtendedChar(b) {
		return s.makeToken(TokenTypeExtendedChar), nil
	}

	if isByteCTL(b) {
		if b == '\r' {
			return s.makeToken(TokenTypeCR), nil
		} else if b == '\n' {
			return s.makeToken(TokenTypeLF), nil
		}

		return s.makeToken(TokenTypeCTL), nil
	}

	switch b {
	case ' ':
		return s.makeToken(TokenTypeSP), nil
	case '!':
		return s.makeToken(TokenTypeExclamation), nil
	case '"':
		return s.makeToken(TokenTypeDQuote), nil
	case '#':
		return s.makeToken(TokenTypeHash), nil
	case '$':
		return s.makeToken(TokenTypeDollar), nil
	case '%':
		return s.makeToken(TokenTypePercent), nil
	case '&':
		return s.makeToken(TokenTypeAmpersand), nil
	case '\'':
		return s.makeToken(TokenTypeSQuote), nil
	case '\\':
		return s.makeToken(TokenTypeBackslash), nil
	case '(':
		return s.makeToken(TokenTypeLParen), nil
	case ')':
		return s.makeToken(TokenTypeRParen), nil
	case '*':
		return s.makeToken(TokenTypeAsterisk), nil
	case '+':
		return s.makeToken(TokenTypePlus), nil
	case ',':
		return s.makeToken(TokenTypeComma), nil
	case '-':
		return s.makeToken(TokenTypeMinus), nil
	case '.':
		return s.makeToken(TokenTypePeriod), nil
	case '/':
		return s.makeToken(TokenTypeSlash), nil
	case ':':
		return s.makeToken(TokenTypeColon), nil
	case ';':
		return s.makeToken(TokenTypeSemicolon), nil
	case '<':
		return s.makeToken(TokenTypeLess), nil
	case '=':
		return s.makeToken(TokenTypeEqual), nil
	case '>':
		return s.makeToken(TokenTypeGreater), nil
	case '?':
		return s.makeToken(TokenTypeQuestion), nil
	case '@':
		return s.makeToken(TokenTypeAt), nil
	case '[':
		return s.makeToken(TokenTypeLBracket), nil
	case ']':
		return s.makeToken(TokenTypeRBracket), nil
	case '^':
		return s.makeToken(TokenTypeCaret), nil
	case '_':
		return s.makeToken(TokenTypeUnderscore), nil
	case '`':
		return s.makeToken(TokenTyeBacktick), nil
	case '{':
		return s.makeToken(TokenTypeLCurly), nil
	case '|':
		return s.makeToken(TokenTypePipe), nil
	case '}':
		return s.makeToken(TokenTypeRCurly), nil
	case '~':
		return s.makeToken(TokenTypeTilde), nil
	}

	return Token{}, fmt.Errorf("unexpected character %v", b)
}

func (s *Scanner) ResetOffsetCounter() {
	s.offset = 0
}

func (s *Scanner) advance() (byte, error) {
	b, err := s.source.ReadByte()
	if err != nil {
		return 0, err
	}

	s.currentByte = b
	s.offset += 1

	return b, nil
}

func (s *Scanner) makeToken(t TokenType) Token {
	return Token{
		TType:  t,
		Value:  s.currentByte,
		Offset: s.offset,
	}
}

func (s *Scanner) makeEOF() Token {
	return Token{
		TType:  TokenTypeEOF,
		Value:  0,
		Offset: s.offset,
	}
}

func isByteAlpha(b byte) bool {
	return (b >= 65 && b <= 90) || (b >= 97 && b <= 122)
}

func isByteDigit(b byte) bool {
	return b >= byte('0') && b <= byte('9')
}

func isByteExtendedChar(b byte) bool {
	return b >= 128
}

func isByteCTL(b byte) bool {
	return b <= 31
}

func ByteToLower(b byte) byte {
	if b >= 65 && b <= 90 {
		return 97 + (b - byte(65))
	}

	return b
}
