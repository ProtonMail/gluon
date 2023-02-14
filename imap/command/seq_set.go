package command

import (
	"fmt"
	"github.com/ProtonMail/gluon/imap/parser"
)

const SeqNumValueAsterisk = SeqNum(0)

type SeqNum int

func (s SeqNum) IsAsterisk() bool {
	return s == SeqNumValueAsterisk
}

func (s SeqNum) String() string {
	if s.IsAsterisk() {
		return "*"
	}

	return fmt.Sprintf("%v", int(s))
}

type SeqRange struct {
	Begin SeqNum
	End   SeqNum
}

func (s SeqRange) String() string {
	return fmt.Sprintf("%v:%v", s.Begin.String(), s.End.String())
}

func ParseNZNumber(p *parser.Parser) (int, error) {
	num, err := p.ParseNumber()
	if err != nil {
		return 0, err
	}

	if num <= 0 {
		return 0, p.MakeError("expected non zero number")
	}

	return num, nil
}

func ParseSeqNumber(p *parser.Parser) (SeqNum, error) {
	if ok, err := p.Matches(parser.TokenTypeAsterisk); err != nil {
		return -1, err
	} else if ok {
		return SeqNumValueAsterisk, nil
	}

	num, err := ParseNZNumber(p)
	if err != nil {
		return -1, err
	}

	return SeqNum(num), nil
}

func ParseSeqRange(p *parser.Parser) (SeqRange, error) {
	seqBegin, err := ParseSeqNumber(p)
	if err != nil {
		return SeqRange{}, err
	}

	if ok, err := p.Matches(parser.TokenTypeColon); err != nil {
		return SeqRange{}, err
	} else if !ok {
		return SeqRange{
			Begin: seqBegin,
			End:   seqBegin,
		}, nil
	}

	seqEnd, err := ParseSeqNumber(p)
	if err != nil {
		return SeqRange{}, err
	}

	return SeqRange{
		Begin: seqBegin,
		End:   seqEnd,
	}, nil
}

func ParseSeqSet(p *parser.Parser) ([]SeqRange, error) {
	// sequence-set    = (seq-number / seq-range) *("," sequence-set)
	var result []SeqRange

	{
		firstRange, err := ParseSeqRange(p)
		if err != nil {
			return nil, err
		}

		result = append(result, firstRange)
	}

	for {
		if ok, err := p.Matches(parser.TokenTypeComma); err != nil {
			return nil, err
		} else if !ok {
			break
		}

		next, err := ParseSeqRange(p)
		if err != nil {
			return nil, err
		}

		result = append(result, next)
	}

	return result, nil
}
