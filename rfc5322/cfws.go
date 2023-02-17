package rfc5322

import "github.com/ProtonMail/gluon/rfcparser"

// Section 3.2.2 White space and Comments

func tryParseCFWS(p *rfcparser.Parser) (bool, error) {
	if !p.CheckWith(func(tokenType rfcparser.TokenType) bool {
		return isWSP(tokenType) || tokenType == rfcparser.TokenTypeCR || tokenType == rfcparser.TokenTypeLParen
	}) {
		return false, nil
	}

	return true, parseCFWS(p)
}

func parseCFWS(p *rfcparser.Parser) error {
	// CFWS            =   (1*([FWS] comment) [FWS]) / FWS
	parsedFirstFWS, err := tryParseFWS(p)
	if err != nil {
		return err
	}

	// Handle case where it can just be FWS without comment
	if !p.Check(rfcparser.TokenTypeLParen) {
		if !parsedFirstFWS {
			return p.MakeError("expected FWS or comment for CFWS")
		}

		return nil
	}

	if err := parseComment(p); err != nil {
		return err
	}

	// Read remaining [FWS] comment
	for {
		if _, err := tryParseFWS(p); err != nil {
			return err
		}

		if !p.Check(rfcparser.TokenTypeLParen) {
			break
		}

		if err := parseComment(p); err != nil {
			return err
		}
	}

	if _, err := tryParseFWS(p); err != nil {
		return err
	}

	return nil
}

func tryParseFWS(p *rfcparser.Parser) (bool, error) {
	if !p.CheckWith(func(tokenType rfcparser.TokenType) bool {
		return isWSP(tokenType) || tokenType == rfcparser.TokenTypeCR
	}) {
		return false, nil
	}

	return true, parseFWS(p)
}

func parseFWS(p *rfcparser.Parser) error {
	// FWS             =   ([*WSP CRLF] 1*WSP) /  obs-FWS
	//                     ; Folding white space
	// obs-FWS         =   1*WSP *(CRLF 1*WSP)
	//
	// Parse 0 or more WSP
	for {
		if ok, err := p.MatchesWith(isWSP); err != nil {
			return err
		} else if !ok {
			break
		}
	}

	if !p.Check(rfcparser.TokenTypeCR) {
		// Early exit.
		return nil
	}

	if err := p.ConsumeNewLine(); err != nil {
		return err
	}

	// Parse one or many WSP.
	if err := p.ConsumeWith(isWSP, "expected WSP after CRLF"); err != nil {
		return err
	}

	for {
		if ok, err := p.MatchesWith(isWSP); err != nil {
			return err
		} else if !ok {
			break
		}
	}

	// Handle obs-FWS case where there can be multiple repeating loops
	for {
		if !p.Check(rfcparser.TokenTypeCR) {
			break
		}

		if err := p.ConsumeNewLine(); err != nil {
			return err
		}

		// Parse one or many WSP.
		if err := p.ConsumeWith(isWSP, "expected WSP after CRLF"); err != nil {
			return err
		}

		for {
			if ok, err := p.MatchesWith(isWSP); err != nil {
				return err
			} else if !ok {
				break
			}
		}
	}

	return nil
}

func parseCContent(p *rfcparser.Parser) error {
	if ok, err := p.MatchesWith(isCText); err != nil {
		return err
	} else if ok {
		return nil
	}

	if _, ok, err := tryParseQuotedPair(p); err != nil {
		return err
	} else if ok {
		return nil
	}

	if p.Check(rfcparser.TokenTypeLParen) {
		return parseComment(p)
	}

	return p.MakeError("unexpected ccontent token")
}

func parseComment(p *rfcparser.Parser) error {
	if err := p.Consume(rfcparser.TokenTypeLParen, "expected ( for comment start"); err != nil {
		return err
	}

	for {
		if _, err := tryParseFWS(p); err != nil {
			return err
		}

		if !p.CheckWith(func(tokenType rfcparser.TokenType) bool {
			return isCText(tokenType) || tokenType == rfcparser.TokenTypeBackslash || tokenType == rfcparser.TokenTypeLParen
		}) {
			break
		}

		if err := parseCContent(p); err != nil {
			return err
		}
	}

	if _, err := tryParseFWS(p); err != nil {
		return err
	}

	if err := p.Consume(rfcparser.TokenTypeRParen, "expected ) for comment end"); err != nil {
		return err
	}

	return nil
}

func tryParseQuotedPair(p *rfcparser.Parser) (byte, bool, error) {
	if !p.Check(rfcparser.TokenTypeBackslash) {
		return 0, false, nil
	}

	b, err := parseQuotedPair(p)
	if err != nil {
		return 0, false, err
	}

	return b, true, nil
}

func parseQuotedPair(p *rfcparser.Parser) (byte, error) {
	// quoted-pair     =   ("\" (VCHAR / WSP)) / obs-qp
	//
	// obs-qp          =   "\" (%d0 / obs-NO-WS-CTL / LF / CR)
	//
	if err := p.Consume(rfcparser.TokenTypeBackslash, "expected \\ for quoted pair start"); err != nil {
		return 0, err
	}

	if ok, err := p.MatchesWith(isVChar); err != nil {
		return 0, err
	} else if ok {
		return p.PreviousToken().Value, nil
	}

	if ok, err := p.MatchesWith(isWSP); err != nil {
		return 0, err
	} else if ok {
		return p.PreviousToken().Value, nil
	}

	if ok, err := p.MatchesWith(func(tokenType rfcparser.TokenType) bool {
		return isObsNoWSCTL(tokenType) ||
			tokenType == rfcparser.TokenTypeCR ||
			tokenType == rfcparser.TokenTypeLF ||
			tokenType == rfcparser.TokenTypeZero
	}); err != nil {
		return 0, err
	} else if ok {
		return p.PreviousToken().Value, nil
	}

	return 0, p.MakeError("unexpected character for quoted pair")
}

func isWSP(tokenType rfcparser.TokenType) bool {
	return tokenType == rfcparser.TokenTypeSP || tokenType == rfcparser.TokenTypeTab
}

func isCText(tokenType rfcparser.TokenType) bool {
	//  ctext           =   %d33-39 /          ; Printable US-ASCII
	//                      %d42-91 /          ;  characters not including
	//                      %d93-126 /         ;  "(", ")", or "\"
	//                      obs-ctext
	//
	//  obs-NO-WS-CTL   =   %d1-8 /            ; US-ASCII control
	//                        %d11 /             ;  characters that do not
	//                        %d12 /             ;  include the carriage
	//                        %d14-31 /          ;  return, line feed, and
	//                        %d127              ;  white space characters
	//
	//  obs-ctext       =   obs-NO-WS-CTL
	switch tokenType { // nolint:exhaustive
	case rfcparser.TokenTypeEOF:
		fallthrough
	case rfcparser.TokenTypeError:
		fallthrough
	case rfcparser.TokenTypeLParen:
		fallthrough
	case rfcparser.TokenTypeRParen:
		fallthrough
	case rfcparser.TokenTypeCR:
		fallthrough
	case rfcparser.TokenTypeTab:
		fallthrough
	case rfcparser.TokenTypeLF:
		fallthrough
	case rfcparser.TokenTypeSP:
		fallthrough
	case rfcparser.TokenTypeBackslash:
		return false
	default:
		return true
	}
}

func isObsNoWSCTL(tokenType rfcparser.TokenType) bool {
	//  obs-NO-WS-CTL   =   %d1-8 /            ; US-ASCII control
	//                        %d11 /             ;  characters that do not
	//                        %d12 /             ;  include the carriage
	//                        %d14-31 /          ;  return, line feed, and
	//                        %d127              ;  white space characters
	switch tokenType { // nolint:exhaustive
	case rfcparser.TokenTypeEOF:
		fallthrough
	case rfcparser.TokenTypeError:
		fallthrough
	case rfcparser.TokenTypeCR:
		fallthrough
	case rfcparser.TokenTypeTab:
		fallthrough
	case rfcparser.TokenTypeLF:
		fallthrough
	case rfcparser.TokenTypeSP:
		return false
	default:
		return rfcparser.IsCTL(tokenType) || tokenType == rfcparser.TokenTypeDelete
	}
}

func isVChar(tokenType rfcparser.TokenType) bool {
	// VChar %x21-7E
	if rfcparser.IsCTL(tokenType) ||
		tokenType == rfcparser.TokenTypeDelete ||
		tokenType == rfcparser.TokenTypeError ||
		tokenType == rfcparser.TokenTypeEOF {
		return false
	}

	return true
}
