package command

import (
	"fmt"
	"github.com/ProtonMail/gluon/rfcparser"
)

type IDGet struct{}

func (l IDGet) String() string {
	return fmt.Sprintf("ID")
}

func (l IDGet) SanitizedString() string {
	return l.String()
}

type IDSet struct {
	Values map[string]string
}

func (l IDSet) String() string {
	if len(l.Values) == 0 {
		return "ID"
	}

	return fmt.Sprintf("ID %v", l.Values)
}

func (l IDSet) SanitizedString() string {
	return l.String()
}

type IDCommandParser struct{}

func (IDCommandParser) FromParser(p *rfcparser.Parser) (Payload, error) {
	// nolint:dupword
	// id ::= "ID" SPACE id_params_list
	//     id_params_list ::= "(" #(string SPACE nstring) ")" / nil
	//         ;; list of field value pairs
	if err := p.Consume(rfcparser.TokenTypeSP, "expected space after command"); err != nil {
		return nil, err
	}

	if p.Check(rfcparser.TokenTypeChar) {
		if err := p.ConsumeBytesFold('N', 'I', 'L'); err != nil {
			return nil, err
		}

		return &IDGet{}, nil
	}

	values := make(map[string]string)

	if err := p.Consume(rfcparser.TokenTypeLParen, "expected ( for id values start"); err != nil {
		return nil, err
	}

	for {
		key, ok, err := p.TryParseString()
		if err != nil {
			return nil, err
		} else if !ok {
			break
		}

		if err := p.Consume(rfcparser.TokenTypeSP, "expected space after ID key"); err != nil {
			return nil, err
		}

		value, isNil, err := ParseNString(p)
		if err != nil {
			return nil, err
		}

		if !isNil {
			values[key.Value] = value.Value
		} else {
			values[key.Value] = ""
		}

		if !p.Check(rfcparser.TokenTypeRParen) {
			if err := p.Consume(rfcparser.TokenTypeSP, "expected space after ID value"); err != nil {
				return nil, err
			}
		}
	}

	if err := p.Consume(rfcparser.TokenTypeRParen, "expected ) for id values end"); err != nil {
		return nil, err
	}

	return &IDSet{
		Values: values,
	}, nil
}
