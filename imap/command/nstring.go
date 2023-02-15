package command

import "github.com/ProtonMail/gluon/rfcparser"

// ParseNString pareses a string or NIL. If NIL was parsed the boolean return is set to false.
func ParseNString(p *rfcparser.Parser) (rfcparser.String, bool, error) {
	// nstring = string / nil
	if s, ok, err := p.TryParseString(); err != nil {
		return rfcparser.String{}, false, err
	} else if ok {
		return s, false, nil
	}

	if err := p.ConsumeBytesFold('N', 'I', 'L'); err != nil {
		return rfcparser.String{}, false, err
	}

	return rfcparser.String{}, true, nil
}
