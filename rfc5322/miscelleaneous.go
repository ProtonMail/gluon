package rfc5322

import (
	"github.com/ProtonMail/gluon/rfcparser"
)

// 3.2.5.  Miscellaneous Tokens

func parseWord(p *rfcparser.Parser) (parserString, error) {
	// word            =   atom / quoted-string
	if _, err := tryParseCFWS(p); err != nil {
		return parserString{}, err
	}

	if p.Check(rfcparser.TokenTypeEqual) {
		return parseEncodedAtom(p)
	}

	if p.Check(rfcparser.TokenTypeDQuote) {
		return parseQuotedString(p)
	}

	result, err := parseAtom(p)
	if err != nil {
		return parserString{}, err
	}

	return result, nil
}

func parsePhrase(p *rfcparser.Parser) ([]parserString, error) {
	// nolint:dupword
	// phrase          =   1*word / obs-phrase
	// obs-phrase      =   word *(word / "." / CFWS)
	// This version has been extended to allow '@' to appear in obs-phrase
	word, err := parseWord(p)
	if err != nil {
		return nil, err
	}

	var result = []parserString{word}

	isSep := func(tokenType rfcparser.TokenType) bool {
		return tokenType == rfcparser.TokenTypePeriod || tokenType == rfcparser.TokenTypeAt
	}

	for {
		// check period case
		if ok, err := p.MatchesWith(isSep); err != nil {
			return nil, err
		} else if ok {
			prevToken := p.PreviousToken()
			result = append(result, parserString{
				String: rfcparser.String{
					Value:  string(prevToken.Value),
					Offset: prevToken.Offset,
				},
				Type: parserStringTypeUnspaced,
			})
			continue
		}

		if _, err := tryParseCFWS(p); err != nil {
			return nil, err
		}

		if !(p.CheckWith(isAText) || p.Check(rfcparser.TokenTypeDQuote)) {
			break
		}

		nextWord, err := parseWord(p)
		if err != nil {
			return nil, err
		}

		result = append(result, nextWord)
	}

	return result, nil
}
