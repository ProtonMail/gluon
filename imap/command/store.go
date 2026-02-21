package command

import (
	"fmt"

	"github.com/ProtonMail/gluon/rfcparser"
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

// StoreDataItem distinguishes between standard IMAP FLAGS and Gmail X-GM-LABELS.
type StoreDataItem int

const (
	StoreDataItemFlags       StoreDataItem = iota // Standard IMAP FLAGS
	StoreDataItemGmailLabels                      // Gmail X-GM-LABELS extension
)

type Store struct {
	SeqSet   []SeqRange
	Action   StoreAction
	Flags    []string
	Silent   bool
	DataItem StoreDataItem
}

func (s Store) String() string {
	silentStr := ""
	if s.Silent {
		silentStr = ".SILENT"
	}

	dataItemStr := s.Action.String()
	if s.DataItem == StoreDataItemGmailLabels {
		switch s.Action {
		case StoreActionAddFlags:
			dataItemStr = "+X-GM-LABELS"
		case StoreActionRemFlags:
			dataItemStr = "-X-GM-LABELS"
		default:
			dataItemStr = "X-GM-LABELS"
		}
	}

	return fmt.Sprintf("STORE %v %v%v %v", s.SeqSet, dataItemStr, silentStr, s.Flags)
}

func (s Store) SanitizedString() string {
	return s.String()
}

type StoreCommandParser struct{}

func (StoreCommandParser) FromParser(p *rfcparser.Parser) (Payload, error) {
	//nolint:dupword
	// store           = "STORE" SP sequence-set SP store-att-flags
	// store-att-flags = (["+" / "-"] "FLAGS" [".SILENT"]) SP
	//                  (flag-list / (flag *(SP flag)))
	// Gmail extension:
	// store-att-flags =/ (["+" / "-"] "X-GM-LABELS" [".SILENT"]) SP
	//                   (label-list)
	if err := p.Consume(rfcparser.TokenTypeSP, "expected space after command"); err != nil {
		return nil, err
	}

	seqSet, err := ParseSeqSet(p)
	if err != nil {
		return nil, err
	}

	if err := p.Consume(rfcparser.TokenTypeSP, "expected space after sequence set"); err != nil {
		return nil, err
	}

	var action StoreAction

	if ok, err := p.Matches(rfcparser.TokenTypePlus); err != nil {
		return nil, err
	} else if !ok {
		if ok, err := p.Matches(rfcparser.TokenTypeMinus); err != nil {
			return nil, err
		} else if ok {
			action = StoreActionRemFlags
		} else {
			action = StoreActionSetFlags
		}
	} else {
		action = StoreActionAddFlags
	}

	// Determine data item: FLAGS or X-GM-LABELS.
	// Peek at current byte to decide which path to take.
	currentByte := rfcparser.ByteToLower(p.CurrentToken().Value)

	var dataItem StoreDataItem

	switch currentByte {
	case 'f':
		// Standard FLAGS data item.
		if err := p.ConsumeBytesFold('F', 'L', 'A', 'G', 'S'); err != nil {
			return nil, err
		}

		dataItem = StoreDataItemFlags
	case 'x':
		// Gmail X-GM-LABELS data item.
		// Consume "X-GM-LABELS" character-by-character.
		if err := p.ConsumeBytesFold('X'); err != nil {
			return nil, err
		}

		if err := p.ConsumeBytesFold('-'); err != nil {
			return nil, err
		}

		if err := p.ConsumeBytesFold('G', 'M'); err != nil {
			return nil, err
		}

		if err := p.ConsumeBytesFold('-'); err != nil {
			return nil, err
		}

		if err := p.ConsumeBytesFold('L', 'A', 'B', 'E', 'L', 'S'); err != nil {
			return nil, err
		}

		dataItem = StoreDataItemGmailLabels
	default:
		return nil, p.MakeError("expected FLAGS or X-GM-LABELS")
	}

	var silent bool

	if ok, err := p.Matches(rfcparser.TokenTypePeriod); err != nil {
		return nil, err
	} else if ok {
		if err := p.ConsumeBytesFold('S', 'I', 'L', 'E', 'N', 'T'); err != nil {
			return nil, err
		}

		silent = true
	}

	if err := p.Consume(rfcparser.TokenTypeSP, "expected space after data item"); err != nil {
		return nil, err
	}

	var values []string

	if dataItem == StoreDataItemGmailLabels {
		values, err = parseGmailLabelList(p)
	} else {
		values, err = parseStoreFlags(p)
	}

	if err != nil {
		return nil, err
	}

	return &Store{
		SeqSet:   seqSet,
		Action:   action,
		Flags:    values,
		Silent:   silent,
		DataItem: dataItem,
	}, nil
}

func parseStoreFlags(p *rfcparser.Parser) ([]string, error) {
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
		if ok, err := p.Matches(rfcparser.TokenTypeSP); err != nil {
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

// parseGmailLabelList parses a Gmail label list: ("Label1" "Label With Spaces" Label3)
// Labels can be quoted strings or atoms.
func parseGmailLabelList(p *rfcparser.Parser) ([]string, error) {
	if err := p.Consume(rfcparser.TokenTypeLParen, "expected '(' at start of Gmail label list"); err != nil {
		return nil, err
	}

	var labels []string

	if !p.Check(rfcparser.TokenTypeRParen) {
		label, err := parseGmailLabel(p)
		if err != nil {
			return nil, err
		}

		labels = append(labels, label)

		for {
			if ok, err := p.Matches(rfcparser.TokenTypeSP); err != nil {
				return nil, err
			} else if !ok {
				break
			}

			label, err := parseGmailLabel(p)
			if err != nil {
				return nil, err
			}

			labels = append(labels, label)
		}
	}

	if err := p.Consume(rfcparser.TokenTypeRParen, "expected ')' at end of Gmail label list"); err != nil {
		return nil, err
	}

	return labels, nil
}

// parseGmailLabel parses a single Gmail label, which can be a quoted string or an atom.
func parseGmailLabel(p *rfcparser.Parser) (string, error) {
	// Try quoted string first (handles labels with spaces).
	if p.Check(rfcparser.TokenTypeDQuote) {
		quoted, err := p.ParseQuoted()
		if err != nil {
			return "", err
		}

		return quoted.Value, nil
	}

	// Try backslash-prefixed system label (e.g., \Inbox).
	if hasBackslash, err := p.Matches(rfcparser.TokenTypeBackslash); err != nil {
		return "", err
	} else if hasBackslash {
		atom, err := p.ParseAtom()
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("\\%v", atom), nil
	}

	// Fall back to atom (plain label name without spaces).
	return p.ParseAtom()
}
