package command

import (
	"fmt"
	"github.com/ProtonMail/gluon/imap/parser"
)

type StoreAction int

const (
	StoreActionAddFlags StoreAction = iota
	StoreActionRemFlags
	StoreActionSetFlags
)

func (s StoreAction) String() string {
	switch s {
	case StoreActionAddFlags:
		return "+FLAGS"
	case StoreActionRemFlags:
		return "+FLAGS"
	case StoreActionSetFlags:
		return "FLAGS"
	default:
		return "UNKNOWN"
	}
}

type StoreCommand struct {
	SeqSet []SeqRange
	Action StoreAction
	Flags  []string
	Silent bool
}

func (s StoreCommand) String() string {
	silentStr := ""
	if s.Silent {
		silentStr = ".SILENT"
	}

	return fmt.Sprintf("STORE %v%v %v", s.Action.String(), silentStr, s.Flags)
}

func (s StoreCommand) SanitizedString() string {
	return s.String()
}

type StoreCommandParser struct{}

func (StoreCommandParser) FromParser(p *parser.Parser) (Payload, error) {
	//nolint:dupword
	// store           = "STORE" SP sequence-set SP store-att-flags
	// store-att-flags = (["+" / "-"] "FLAGS" [".SILENT"]) SP
	//                  (flag-list / (flag *(SP flag)))
	if err := p.Consume(parser.TokenTypeSP, "expected space after command"); err != nil {
		return nil, err
	}

	seqSet, err := ParseSeqSet(p)
	if err != nil {
		return nil, err
	}

	if err := p.Consume(parser.TokenTypeSP, "expected space after sequence set"); err != nil {
		return nil, err
	}

	var action StoreAction

	if ok, err := p.Matches(parser.TokenTypePlus); err != nil {
		return nil, err
	} else if !ok {
		if ok, err := p.Matches(parser.TokenTypeMinus); err != nil {
			return nil, err
		} else if ok {
			action = StoreActionRemFlags
		} else {
			action = StoreActionSetFlags
		}
	} else {
		action = StoreActionAddFlags
	}

	if err := p.ConsumeBytesFold('F', 'L', 'A', 'G', 'S'); err != nil {
		return nil, err
	}

	var silent bool

	if ok, err := p.Matches(parser.TokenTypePeriod); err != nil {
		return nil, err
	} else if ok {
		if err := p.ConsumeBytesFold('S', 'I', 'L', 'E', 'N', 'T'); err != nil {
			return nil, err
		}

		silent = true
	}

	if err := p.Consume(parser.TokenTypeSP, "expected space after FLAGS"); err != nil {
		return nil, err
	}

	flags, err := parseStoreFlags(p)
	if err != nil {
		return nil, err
	}

	return &StoreCommand{
		SeqSet: seqSet,
		Action: action,
		Flags:  flags,
		Silent: silent,
	}, nil
}

func parseStoreFlags(p *parser.Parser) ([]string, error) {
	//                  (flag-list / (flag *(SP flag)))
	fl, ok, err := TryParseFlagList(p)
	if err != nil {
		return nil, err
	} else if ok {
		return fl, nil
	}

	var flags []string

	// first flag.
	{
		f, err := ParseFlag(p)
		if err != nil {
			return nil, err
		}

		flags = append(flags, f)
	}

	// remaining.
	for {
		if ok, err := p.Matches(parser.TokenTypeSP); err != nil {
			return nil, err
		} else if !ok {
			break
		}

		f, err := ParseFlag(p)
		if err != nil {
			return nil, err
		}

		flags = append(flags, f)
	}

	return flags, nil
}
