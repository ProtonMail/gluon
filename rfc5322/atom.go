package rfc5322

// 3.2.4.  Quoted Strings

import (
	"fmt"
	"io"
	"mime"

	"github.com/ProtonMail/gluon/rfcparser"
)

func parseDotAtom(p *rfcparser.Parser) (rfcparser.String, error) {
	// dot-atom        =   [CFWS] dot-atom-text [CFWS]
	if _, err := tryParseCFWS(p); err != nil {
		return rfcparser.String{}, err
	}

	atom, err := parseDotAtomText(p)
	if err != nil {
		return rfcparser.String{}, err
	}

	if _, err := tryParseCFWS(p); err != nil {
		return rfcparser.String{}, err
	}

	return atom, nil
}

func parseDotAtomText(p *rfcparser.Parser) (rfcparser.String, error) {
	//  dot-atom-text   =   1*atext *("." 1*atext)
	//  This version has been extended to allow for trailing '.' files.
	if err := p.ConsumeWith(isAText, "expected atext char for dot-atom-text"); err != nil {
		return rfcparser.String{}, err
	}

	atom, err := p.CollectBytesWhileMatchesWithPrevWith(isAText)
	if err != nil {
		return rfcparser.String{}, err
	}

	for {
		if ok, err := p.Matches(rfcparser.TokenTypePeriod); err != nil {
			return rfcparser.String{}, err
		} else if !ok {
			break
		}

		atom.Value = append(atom.Value, '.')

		if p.Check(rfcparser.TokenTypePeriod) {
			return rfcparser.String{}, p.MakeError("invalid token after '.'")
		}

		// Early exit to allow trailing '.'
		if !p.CheckWith(isAText) {
			break
		}

		if err := p.ConsumeWith(isAText, "expected atext char for dot-atom-text"); err != nil {
			return rfcparser.String{}, err
		}

		atomNext, err := p.CollectBytesWhileMatchesWithPrevWith(isAText)
		if err != nil {
			return rfcparser.String{}, err
		}

		atom.Value = append(atom.Value, atomNext.Value...)
	}

	return atom.IntoString(), nil
}

func parseAtom(p *rfcparser.Parser) (parserString, error) {
	// atom            =   [CFWS] 1*atext [CFWS]
	if _, err := tryParseCFWS(p); err != nil {
		return parserString{}, err
	}

	if err := p.ConsumeWith(isAText, "expected atext char for atom"); err != nil {
		return parserString{}, err
	}

	atom, err := p.CollectBytesWhileMatchesWithPrevWith(isAText)
	if err != nil {
		return parserString{}, err
	}

	if _, err := tryParseCFWS(p); err != nil {
		return parserString{}, err
	}

	return parserString{
		String: atom.IntoString(),
		Type:   parserStringTypeOther,
	}, nil
}

var CharsetReader func(charset string, input io.Reader) (io.Reader, error)

func parseEncodedAtom(p *rfcparser.Parser) (parserString, error) {
	// encoded-word = "=?" charset "?" encoding "?" encoded-text "?="
	//
	// charset = token    ; see section 3
	//
	// encoding = token   ; see section 4
	//
	//
	if _, err := tryParseCFWS(p); err != nil {
		return parserString{}, err
	}

	var fullWord string

	startOffset := p.CurrentToken().Offset

	if err := p.ConsumeBytesFold('=', '?'); err != nil {
		return parserString{}, err
	}

	fullWord += "=?"

	charset, err := p.CollectBytesWhileMatchesWith(isEncodedAtomToken)
	if err != nil {
		return parserString{}, err
	}

	fullWord += charset.IntoString().Value

	if err := p.Consume(rfcparser.TokenTypeQuestion, "expected '?' after encoding charset"); err != nil {
		return parserString{}, err
	}

	fullWord += "?"

	if err := p.Consume(rfcparser.TokenTypeChar, "expected char after '?'"); err != nil {
		return parserString{}, err
	}

	encoding := rfcparser.ByteToLower(p.PreviousToken().Value)
	if encoding != 'q' && encoding != 'b' {
		return parserString{}, p.MakeError("encoding should either be 'Q' or 'B'")
	}

	if err := p.Consume(rfcparser.TokenTypeQuestion, "expected '?' after encoding byte"); err != nil {
		return parserString{}, err
	}

	if encoding == 'b' {
		fullWord += "B"
	} else {
		fullWord += "Q"
	}

	fullWord += "?"

	encodedText, err := p.CollectBytesWhileMatchesWith(isEncodedText)
	if err != nil {
		return parserString{}, err
	}

	fullWord += encodedText.IntoString().Value

	if err := p.ConsumeBytesFold('?', '='); err != nil {
		return parserString{}, err
	}

	fullWord += "?="

	if _, err := tryParseCFWS(p); err != nil {
		return parserString{}, err
	}

	decoder := mime.WordDecoder{CharsetReader: CharsetReader}

	decoded, err := decoder.Decode(fullWord)
	if err != nil {
		return parserString{}, p.MakeErrorAtOffset(fmt.Sprintf("failed to decode encoded atom: %v", err), startOffset)
	}

	return parserString{
		String: rfcparser.String{Value: decoded, Offset: startOffset},
		Type:   parserStringTypeEncoded,
	}, nil
}

func isEncodedAtomToken(tokenType rfcparser.TokenType) bool {
	// token = 1*<Any CHAR except SPACE, CTLs, and especials>
	//
	// specials = "(" / ")" / "<" / ">" / "@" / "," / ";" / ":" / "
	// <"> / "/" / "[" / "]" / "?" / "." / "="
	if rfcparser.IsCTL(tokenType) {
		return false
	}

	switch tokenType { //nolint:exhaustive
	case rfcparser.TokenTypeEOF:
		fallthrough
	case rfcparser.TokenTypeError:
		fallthrough
	case rfcparser.TokenTypeSP:
		fallthrough
	case rfcparser.TokenTypeLParen:
		fallthrough
	case rfcparser.TokenTypeRParen:
		fallthrough
	case rfcparser.TokenTypeLess:
		fallthrough
	case rfcparser.TokenTypeGreater:
		fallthrough
	case rfcparser.TokenTypeAt:
		fallthrough
	case rfcparser.TokenTypeComma:
		fallthrough
	case rfcparser.TokenTypeSemicolon:
		fallthrough
	case rfcparser.TokenTypeColon:
		fallthrough
	case rfcparser.TokenTypeDQuote:
		fallthrough
	case rfcparser.TokenTypeSlash:
		fallthrough
	case rfcparser.TokenTypeLBracket:
		fallthrough
	case rfcparser.TokenTypeRBracket:
		fallthrough
	case rfcparser.TokenTypeQuestion:
		fallthrough
	case rfcparser.TokenTypePeriod:
		fallthrough
	case rfcparser.TokenTypeEqual:
		return false
	default:
		return true
	}
}

func isEncodedText(tokenType rfcparser.TokenType) bool {
	//  encoded-text = 1*<Any printable ASCII character other than "?"
	//                     or SPACE>
	//                  ; (but see "Use of encoded-words in message
	//                  ; headers", section 5)
	//
	if rfcparser.IsCTL(tokenType) ||
		tokenType == rfcparser.TokenTypeSP ||
		tokenType == rfcparser.TokenTypeQuestion ||
		tokenType == rfcparser.TokenTypeEOF ||
		tokenType == rfcparser.TokenTypeError ||
		tokenType == rfcparser.TokenTypeExtendedChar {
		return false
	}

	return true
}

func isAText(tokenType rfcparser.TokenType) bool {
	//     atext           =   ALPHA / DIGIT /    ; Printable US-ASCII
	//                         "!" / "#" /        ;  characters not including
	//                         "$" / "%" /        ;  specials.  Used for atoms.
	//                         "&" / "'" /
	//                         "*" / "+" /
	//                         "-" / "/" /
	//                         "=" / "?" /
	//                         "^" / "_" /
	//                         "`" / "{" /
	//                         "|" / "}" /
	//                         "~"
	switch tokenType { //nolint:exhaustive
	case rfcparser.TokenTypeDigit:
		fallthrough
	case rfcparser.TokenTypeChar:
		fallthrough
	case rfcparser.TokenTypeExclamation:
		fallthrough
	case rfcparser.TokenTypeHash:
		fallthrough
	case rfcparser.TokenTypeDollar:
		fallthrough
	case rfcparser.TokenTypePercent:
		fallthrough
	case rfcparser.TokenTypeAmpersand:
		fallthrough
	case rfcparser.TokenTypeSQuote:
		fallthrough
	case rfcparser.TokenTypeAsterisk:
		fallthrough
	case rfcparser.TokenTypePlus:
		fallthrough
	case rfcparser.TokenTypeMinus:
		fallthrough
	case rfcparser.TokenTypeSlash:
		fallthrough
	case rfcparser.TokenTypeEqual:
		fallthrough
	case rfcparser.TokenTypeQuestion:
		fallthrough
	case rfcparser.TokenTypeCaret:
		fallthrough
	case rfcparser.TokenTypeUnderscore:
		fallthrough
	case rfcparser.TokenTyeBacktick:
		fallthrough
	case rfcparser.TokenTypeLCurly:
		fallthrough
	case rfcparser.TokenTypeRCurly:
		fallthrough
	case rfcparser.TokenTypePipe:
		fallthrough
	case rfcparser.TokenTypeExtendedChar: // RFC6532
		fallthrough
	case rfcparser.TokenTypeTilde:
		return true
	default:
		return false
	}
}
