package rfc5322

import (
	"net/mail"

	"github.com/ProtonMail/gluon/rfcparser"
)

// 3.4.  Address Specification

func parseAddressList(p *Parser) ([]*mail.Address, error) {
	// address-list    =   (address *("," address)) / obs-addr-list
	//  *([CFWS] ",") address *("," [address / CFWS])
	// We extended this rule to allow ';' as separator
	var result []*mail.Address

	isSep := func(tokenType rfcparser.TokenType) bool {
		return tokenType == rfcparser.TokenTypeComma || tokenType == rfcparser.TokenTypeSemicolon
	}

	// *([CFWS] ",")
	for {
		if _, err := tryParseCFWS(p.parser); err != nil {
			return nil, err
		}

		if ok, err := p.parser.MatchesWith(isSep); err != nil {
			return nil, err
		} else if !ok {
			break
		}
	}

	var groupConsumedSemiColon bool
	// Address
	{
		addr, gConsumedSemiColon, err := parseAddress(p)
		if err != nil {
			return nil, err
		}

		groupConsumedSemiColon = gConsumedSemiColon

		result = append(result, addr...)
	}

	// *("," [address / CFWS])
	for {
		if ok, err := p.parser.MatchesWith(isSep); err != nil {
			return nil, err
		} else if !ok { // see `parseAddress` comment about why this is necessary.
			if !groupConsumedSemiColon || p.parser.CurrentToken().TType == rfcparser.TokenTypeEOF {
				break
			}
		}

		if ok, err := tryParseCFWS(p.parser); err != nil {
			return nil, err
		} else if ok {
			// Only continue if the next input is EOF or comma or we can run into issues with parsring
			// the `',' address` rules.
			if p.parser.Check(rfcparser.TokenTypeEOF) || p.parser.CheckWith(isSep) {
				continue
			}
		}

		// address
		addr, consumedSemiColon, err := parseAddress(p)
		if err != nil {
			return nil, err
		}

		groupConsumedSemiColon = consumedSemiColon

		result = append(result, addr...)
	}

	return result, nil
}

// The boolean parameter represents whether a group consumed a ';' separator. This is necessary to disambiguate
// an address list where we have the sequence ` g:<address>;<address>` since we also allow groups to have optional
// `;` terminators.
func parseAddress(p *Parser) ([]*mail.Address, bool, error) {
	//    address         =   mailbox / group
	//    name-addr       =   [display-name] angle-addr
	//    group           =   display-name ":" [group-list] ";" [CFWS]
	//
	if _, err := tryParseCFWS(p.parser); err != nil {
		return nil, false, err
	}

	// check addr-spec standalone
	if p.parser.Check(rfcparser.TokenTypeLess) {
		addr, err := parseAngleAddr(p.parser)
		if err != nil {
			return nil, false, err
		}

		return []*mail.Address{{
			Name:    "",
			Address: addr,
		}}, false, nil
	}

	parserState := p.SaveState()

	if address, err := parseMailbox(p); err == nil {
		return []*mail.Address{
			address,
		}, false, nil
	}

	p.RestoreState(parserState)

	group, didConsumeSemicolon, err := parseGroup(p)
	if err != nil {
		return nil, false, err
	}

	return group, didConsumeSemicolon, nil
}

func parseGroup(p *Parser) ([]*mail.Address, bool, error) {
	// nolint:dupword
	// group           =   display-name ":" [group-list] ";" [CFWS]
	// group-list      =   mailbox-list / CFWS / obs-group-list
	// obs-group-list  =   1*([CFWS] ",") [CFWS]
	//
	// nolint:dupword
	// mailbox-list    =   (mailbox *("," mailbox)) / obs-mbox-list
	// obs-mbox-list   =   *([CFWS] ",") mailbox *("," [mailbox / CFWS])
	//
	// This version has been relaxed so that the ';' is optional. and that a group can be wrapped in `"`
	hasQuotes, err := p.parser.Matches(rfcparser.TokenTypeDQuote)
	if err != nil {
		return nil, false, err
	}

	if _, err := parseDisplayName(p.parser); err != nil {
		return nil, false, err
	}

	if err := p.parser.Consume(rfcparser.TokenTypeColon, "expected ':' for group start"); err != nil {
		return nil, false, err
	}

	var didConsumeSemicolon bool

	var result []*mail.Address

	if ok, err := p.parser.Matches(rfcparser.TokenTypeSemicolon); err != nil {
		return nil, false, err
	} else if !ok {

		// *([CFWS] ",")
		for {
			if _, err := tryParseCFWS(p.parser); err != nil {
				return nil, false, err
			}

			if ok, err := p.parser.Matches(rfcparser.TokenTypeComma); err != nil {
				return nil, false, err
			} else if !ok {
				break
			}
		}

		// This section can optionally be one of the following: mailbox-list / CFWS / obs-group-list. So if
		// we run out of input, we see semicolon or a double quote we should skip trying to parse this bit.
		if !(p.parser.Check(rfcparser.TokenTypeEOF) ||
			p.parser.Check(rfcparser.TokenTypeSemicolon) ||
			p.parser.Check(rfcparser.TokenTypeDQuote)) {
			// Mailbox
			var parsedFirstMailbox bool

			{
				parserState := p.SaveState()
				mailbox, err := parseMailbox(p)
				if err != nil {
					p.RestoreState(parserState)
				} else {
					parsedFirstMailbox = true
					result = append(result, mailbox)
				}
			}

			// *("," [mailbox / CFWS])
			if parsedFirstMailbox {
				for {
					if ok, err := p.parser.Matches(rfcparser.TokenTypeComma); err != nil {
						return nil, false, err
					} else if !ok {
						break
					}

					if ok, err := tryParseCFWS(p.parser); err != nil {
						return nil, false, err
					} else if ok {
						continue
					}

					// Mailbox
					mailbox, err := parseMailbox(p)
					if err != nil {
						return nil, false, err
					}

					result = append(result, mailbox)
				}
			} else {
				// If we did not parse a mailbox then we must parse CWFS
				if err := parseCFWS(p.parser); err != nil {
					return nil, false, err
				}
			}
		}

		consumedSemicolon, err := p.parser.Matches(rfcparser.TokenTypeSemicolon)
		if err != nil {
			return nil, false, err
		}

		didConsumeSemicolon = consumedSemicolon
	} else {
		didConsumeSemicolon = true
	}

	if _, err := tryParseCFWS(p.parser); err != nil {
		return nil, false, err
	}

	if hasQuotes {
		if err := p.parser.Consume(rfcparser.TokenTypeDQuote, `expected '"' for group end`); err != nil {
			return nil, false, err
		}
	}

	return result, didConsumeSemicolon, nil
}

func parseMailbox(p *Parser) (*mail.Address, error) {
	//    mailbox         =   name-addr / addr-spec
	parserState := p.SaveState()

	if addr, err := parseNameAddr(p.parser); err == nil {
		return addr, nil
	}

	p.RestoreState(parserState)

	addr, err := parseAddrSpec(p.parser)
	if err != nil {
		return nil, err
	}

	return &mail.Address{
		Address: addr,
	}, nil
}

func parseNameAddr(p *rfcparser.Parser) (*mail.Address, error) {
	// name-addr       =   [display-name] angle-addr
	if _, err := tryParseCFWS(p); err != nil {
		return nil, err
	}

	// Only has angle-addr component.
	if p.Check(rfcparser.TokenTypeLess) {
		address, err := parseAngleAddr(p)
		if err != nil {
			return nil, err
		}

		return &mail.Address{Address: address}, nil
	}

	displayName, err := parseDisplayName(p)
	if err != nil {
		return nil, err
	}

	address, err := parseAngleAddr(p)
	if err != nil {
		return nil, err
	}

	return &mail.Address{Address: address, Name: displayName}, nil
}

func parseAngleAddr(p *rfcparser.Parser) (string, error) {
	// angle-addr      =   [CFWS] "<" addr-spec ">" [CFWS] /
	//                        obs-angle-addr
	//
	//      obs-angle-addr  =   [CFWS] "<" obs-route addr-spec ">" [CFWS]
	//
	//      obs-route       =   obs-domain-list ":"
	//
	//      obs-domain-list =   *(CFWS / ",") "@" domain
	//                          *("," [CFWS] ["@" domain])
	//
	// This version has been extended so that add-rspec is optional
	if _, err := tryParseCFWS(p); err != nil {
		return "", err
	}

	if err := p.Consume(rfcparser.TokenTypeLess, "expected < for angle-addr start"); err != nil {
		return "", err
	}

	if ok, err := p.Matches(rfcparser.TokenTypeGreater); err != nil {
		return "", err
	} else if ok {
		return "", nil
	}

	for {
		if ok, err := tryParseCFWS(p); err != nil {
			return "", err
		} else if !ok {
			if ok, err := p.Matches(rfcparser.TokenTypeComma); err != nil {
				return "", err
			} else if !ok {
				break
			}
		}
	}

	if ok, err := p.Matches(rfcparser.TokenTypeAt); err != nil {
		return "", err
	} else if ok {
		if _, err := parseDomain(p); err != nil {
			return "", err
		}

		for {
			if ok, err := p.Matches(rfcparser.TokenTypeComma); err != nil {
				return "", err
			} else if !ok {
				break
			}

			if _, err := tryParseCFWS(p); err != nil {
				return "", err
			}

			if ok, err := p.Matches(rfcparser.TokenTypeAt); err != nil {
				return "", err
			} else if ok {
				if _, err := parseDomain(p); err != nil {
					return "", err
				}
			}
		}

		if err := p.Consume(rfcparser.TokenTypeColon, "expected ':' for obs-route end"); err != nil {
			return "", err
		}
	}

	addr, err := parseAddrSpec(p)
	if err != nil {
		return "", err
	}

	if err := p.Consume(rfcparser.TokenTypeGreater, "expected > for angle-addr end"); err != nil {
		return "", err
	}

	if _, err := tryParseCFWS(p); err != nil {
		return "", err
	}

	return addr, nil
}

func parseDisplayName(p *rfcparser.Parser) (string, error) {
	// display-name    =   phrase
	phrase, err := parsePhrase(p)
	if err != nil {
		return "", err
	}

	return joinWithSpacingRules(phrase), nil
}

func parseAddrSpec(p *rfcparser.Parser) (string, error) {
	//     addr-spec       =   local-part "@" domain
	// This version adds an option port extension : COLON ATOM
	localPart, err := parseLocalPart(p)
	if err != nil {
		return "", err
	}

	if err := p.Consume(rfcparser.TokenTypeAt, "expected @ after local-part"); err != nil {
		return "", err
	}

	domain, err := parseDomain(p)
	if err != nil {
		return "", err
	}

	if ok, err := p.Matches(rfcparser.TokenTypeColon); err != nil {
		return "", err
	} else if ok {
		port, err := parseAtom(p)
		if err != nil {
			return "", err
		}

		return localPart + "@" + domain + ":" + port.String.Value, nil
	}

	return localPart + "@" + domain, nil
}

func parseLocalPart(p *rfcparser.Parser) (string, error) {
	// nolint:dupword
	//     local-part      =   dot-atom / quoted-string / obs-local-part
	// 	   obs-local-part  =   word *("." word)
	//     word            =   atom / quoted-string
	// ^ above rule can be relaxed into just the last part, dot-atom just
	// Local part extended
	var words []parserString

	{
		word, err := parseWord(p)
		if err != nil {
			return "", err
		}

		words = append(words, word)
	}

	for {
		if ok, err := p.Matches(rfcparser.TokenTypePeriod); err != nil {
			return "", err
		} else if !ok {
			break
		}

		words = append(words, parserString{
			String: rfcparser.String{
				Value:  ".",
				Offset: p.PreviousToken().Offset,
			},
			Type: parserStringTypeUnspaced,
		})

		word, err := parseWord(p)
		if err != nil {
			return "", err
		}

		words = append(words, word)
	}

	return joinWithSpacingRules(words), nil
}

func parseDomain(p *rfcparser.Parser) (string, error) {
	//     domain          =   dot-atom / domain-literal / obs-domain
	//
	//     obs-domain      =   atom *("." atom)
	//
	if _, err := tryParseCFWS(p); err != nil {
		return "", err
	}

	if ok, err := p.Matches(rfcparser.TokenTypeLBracket); err != nil {
		return "", err
	} else if ok {
		return parseDomainLiteral(p)
	}

	// obs-domain can be seen as a more restrictive dot-atom so we just use that rule instead.
	dotAtom, err := parseDotAtom(p)
	if err != nil {
		return "", err
	}

	return dotAtom.Value, nil
}

func parseDomainLiteral(p *rfcparser.Parser) (string, error) {
	//     domain-literal  =   [CFWS] "[" *([FWS] dtext) [FWS] "]" [CFWS]
	//
	// [CFWS] and "[" consumed before entry
	//
	result := []byte{'['}

	for {
		if _, err := tryParseFWS(p); err != nil {
			return "", err
		}

		if ok, err := p.MatchesWith(isDText); err != nil {
			return "", err
		} else if !ok {
			break
		}

		result = append(result, p.PreviousToken().Value)
	}

	if _, err := tryParseFWS(p); err != nil {
		return "", err
	}

	if err := p.Consume(rfcparser.TokenTypeRBracket, "expecetd ] for domain-literal end"); err != nil {
		return "", err
	}

	result = append(result, ']')

	if _, err := tryParseCFWS(p); err != nil {
		return "", err
	}

	return string(result), nil
}

func isDText(tokenType rfcparser.TokenType) bool {
	//     dtext           =   %d33-90 /          ; Printable US-ASCII
	//                         %d94-126 /         ;  characters not including
	//                         obs-dtext          ;  "[", "]", or "\"
	//
	//     obs-dtext       =   obs-NO-WS-CTL / quoted-pair // <- we have not included this
	//
	if rfcparser.IsCTL(tokenType) ||
		tokenType == rfcparser.TokenTypeLBracket ||
		tokenType == rfcparser.TokenTypeRBracket ||
		tokenType == rfcparser.TokenTypeBackslash {
		return false
	}

	return true
}

func joinWithSpacingRules(v []parserString) string {
	result := v[0].String.Value

	prevStrType := v[0].Type

	for i := 1; i < len(v); i++ {
		curStrType := v[i].Type

		if prevStrType == parserStringTypeEncoded {
			if curStrType == parserStringTypeOther {
				result += " "
			}
		} else if prevStrType != parserStringTypeUnspaced {
			if curStrType != parserStringTypeUnspaced {
				result += " "
			}
		}

		prevStrType = curStrType

		result += v[i].String.Value
	}

	return result
}
