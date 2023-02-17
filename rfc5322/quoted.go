package rfc5322

// 3.2.4.  Quoted Strings

import "github.com/ProtonMail/gluon/rfcparser"

func parseQuotedString(p *rfcparser.Parser) (parserString, error) {
	var result rfcparser.Bytes
	result.Offset = p.CurrentToken().Offset

	if _, err := tryParseCFWS(p); err != nil {
		return parserString{}, err
	}

	if err := p.Consume(rfcparser.TokenTypeDQuote, "expected \" for quoted string start"); err != nil {
		return parserString{}, err
	}

	for {
		if ok, err := tryParseFWS(p); err != nil {
			return parserString{}, err
		} else if ok {
			result.Value = append(result.Value, ' ')
		}

		if !(p.CheckWith(isQText) || p.Check(rfcparser.TokenTypeBackslash)) {
			break
		}

		if p.CheckWith(isQText) {
			b, err := parseQContent(p)
			if err != nil {
				return parserString{}, err
			}

			result.Value = append(result.Value, b)
		} else {
			b, err := parseQuotedPair(p)
			if err != nil {
				return parserString{}, err
			}

			result.Value = append(result.Value, b)
		}
	}

	if ok, err := tryParseFWS(p); err != nil {
		return parserString{}, err
	} else if ok {
		result.Value = append(result.Value, ' ')
	}

	if err := p.Consume(rfcparser.TokenTypeDQuote, "expected \" for quoted string end"); err != nil {
		return parserString{}, err
	}

	if _, err := tryParseCFWS(p); err != nil {
		return parserString{}, err
	}

	return parserString{
		String: result.IntoString(),
		Type:   parserStringTypeOther,
	}, nil
}

func parseQContent(p *rfcparser.Parser) (byte, error) {
	if ok, err := p.MatchesWith(isQText); err != nil {
		return 0, err
	} else if ok {
		return p.PreviousToken().Value, nil
	}

	return parseQuotedPair(p)
}

func isQText(tokenType rfcparser.TokenType) bool {
	//     qtext           =   %d33 /             ; Printable US-ASCII
	//                         %d35-91 /          ;  characters not including
	//                         %d93-126 /         ;  "\" or the quote character
	//                         obs-qtext
	//
	// 	obs-qtext       =   obs-NO-WS-CTL
	//
	if (rfcparser.IsCTL(tokenType) && !isObsNoWSCTL(tokenType)) ||
		tokenType == rfcparser.TokenTypeDQuote ||
		tokenType == rfcparser.TokenTypeBackslash ||
		tokenType == rfcparser.TokenTypeSP ||
		tokenType == rfcparser.TokenTypeEOF ||
		tokenType == rfcparser.TokenTypeError {
		return false
	}

	return true
}
