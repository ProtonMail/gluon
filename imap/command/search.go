package command

import (
	"fmt"
	rfcparser "github.com/ProtonMail/gluon/rfcparser"
	"github.com/bradenaw/juniper/xslices"
	"time"
)

type SearchCommand struct {
	Charset string
	Keys    []SearchKey
}

type SearchKey interface {
	String() string
	SanitizedString() string
}

func (s SearchCommand) String() string {
	charsetStr := "NONE"
	if len(s.Charset) != 0 {
		charsetStr = s.Charset
	}

	return fmt.Sprintf("SEARCH CHARSET=%v %v", charsetStr, s.Keys)
}

func (s SearchCommand) SanitizedString() string {
	charsetStr := "NONE"
	if len(s.Charset) != 0 {
		charsetStr = s.Charset
	}

	return fmt.Sprintf("SEARCH CHARSET=%v %v", charsetStr, xslices.Map(s.Keys, func(v SearchKey) string {
		return v.SanitizedString()
	}))
}

type SearchCommandParser struct{}

func (scp *SearchCommandParser) FromParser(p *rfcparser.Parser) (Payload, error) {
	//search          = "SEARCH" [SP "CHARSET" SP astring] 1*(SP search-key)
	//                     ; CHARSET argument to MUST be registered with IANA
	var keys []SearchKey

	var charset string

	// check for optional charset
	keyword, err := readSearchKeyword(p)
	if err != nil {
		return nil, err
	}

	if keyword.Value == "charset" {
		if err := p.Consume(rfcparser.TokenTypeSP, "expected space after charset"); err != nil {
			return nil, err
		}

		encoding, err := p.ParseAString()
		if err != nil {
			return nil, err
		}

		charset = encoding.Value
	} else {
		// Not charset, perform handling of the keword
		key, err := handleSearchKey(keyword, p)
		if err != nil {
			return nil, err
		}

		keys = append(keys, key)
	}

	for {
		if !p.Check(rfcparser.TokenTypeSP) {
			break
		}

		key, err := parseSearchKey(p)
		if err != nil {
			return nil, err
		}

		keys = append(keys, key)
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("no search keys specified")
	}

	return &SearchCommand{
		Charset: charset,
		Keys:    keys,
	}, nil
}

func parseSearchKey(p *rfcparser.Parser) (SearchKey, error) {
	keyword, err := readSearchKeyword(p)
	if err != nil {
		return nil, err
	}

	return handleSearchKey(keyword, p)
}

func readSearchKeyword(p *rfcparser.Parser) (rfcparser.String, error) {
	if err := p.Consume(rfcparser.TokenTypeSP, "expected space"); err != nil {
		return rfcparser.String{}, err
	}

	keyword, err := p.CollectBytesWhileMatches(rfcparser.TokenTypeChar)
	if err != nil {
		return rfcparser.String{}, err
	}

	return keyword.IntoString().ToLower(), nil
}

func handleSearchKey(keyword rfcparser.String, p *rfcparser.Parser) (SearchKey, error) {
	/*
	  search-key      = "ALL" / "ANSWERED" / "BCC" SP astring /
	                    "BEFORE" SP date / "BODY" SP astring /
	                    "CC" SP astring / "DELETED" / "FLAGGED" /
	                    "FROM" SP astring / "KEYWORD" SP flag-keyword /
	                    "NEW" / "OLD" / "ON" SP date / "RECENT" / "SEEN" /
	                    "SINCE" SP date / "SUBJECT" SP astring /
	                    "TEXT" SP astring / "TO" SP astring /
	                    "UNANSWERED" / "UNDELETED" / "UNFLAGGED" /
	                    "UNKEYWORD" SP flag-keyword / "UNSEEN" /
	                      ; Above this line were in [IMAP2]
	                    "DRAFT" / "HEADER" SP header-fld-name SP astring /
	                    "LARGER" SP number / "NOT" SP search-key /
	                    "OR" SP search-key SP search-key /
	                    "SENTBEFORE" SP date / "SENTON" SP date /
	                    "SENTSINCE" SP date / "SMALLER" SP number /
	                    "UID" SP sequence-set / "UNDRAFT" / sequence-set /
	                    "(" search-key *(SP search-key) ")"     	*/
	switch keyword.Value {
	case "all":
		return &SearchKeyAll{}, nil

	case "answered":
		return &SearchKeyAnswered{}, nil

	case "bcc":
		value, err := parseStringKeyAString(p)
		if err != nil {
			return nil, err
		}

		return &SearchKeyBCC{Value: value}, nil

	case "before":
		value, err := parseStringKeyDate(p)
		if err != nil {
			return nil, err
		}

		return &SearchKeyBefore{Value: value}, nil

	case "on":
		value, err := parseStringKeyDate(p)
		if err != nil {
			return nil, err
		}

		return &SearchKeyOn{Value: value}, nil

	case "body":
		value, err := parseStringKeyAString(p)
		if err != nil {
			return nil, err
		}

		return &SearchKeyBody{Value: value}, nil

	case "cc":
		value, err := parseStringKeyAString(p)
		if err != nil {
			return nil, err
		}

		return &SearchKeyCC{Value: value}, nil

	case "deleted":
		return &SearchKeyDeleted{}, nil

	case "flagged":
		return &SearchKeyFlagged{}, nil

	case "from":
		value, err := parseStringKeyAString(p)
		if err != nil {
			return nil, err
		}

		return &SearchKeyFrom{Value: value}, nil

	case "keyword":
		value, err := parseStringKeyAtom(p)
		if err != nil {
			return nil, err
		}

		return &SearchKeyKeyword{Value: value}, nil

	case "new":
		return &SearchKeyNew{}, nil

	case "old":
		return &SearchKeyOld{}, nil

	case "recent":
		return &SearchKeyRecent{}, nil

	case "seen":
		return &SearchKeySeen{}, nil

	case "since":
		value, err := parseStringKeyDate(p)
		if err != nil {
			return nil, err
		}

		return &SearchKeySince{Value: value}, nil

	case "subject":
		value, err := parseStringKeyAString(p)
		if err != nil {
			return nil, err
		}

		return &SearchKeySubject{Value: value}, nil

	case "text":
		value, err := parseStringKeyAString(p)
		if err != nil {
			return nil, err
		}

		return &SearchKeyText{Value: value}, nil

	case "to":
		value, err := parseStringKeyAString(p)
		if err != nil {
			return nil, err
		}

		return &SearchKeyTo{Value: value}, nil

	case "unanswered":
		return &SearchKeyUnanswered{}, nil

	case "undeleted":
		return &SearchKeyUndeleted{}, nil

	case "unflagged":
		return &SearchKeyUnflagged{}, nil

	case "unkeyword":
		value, err := parseStringKeyAtom(p)
		if err != nil {
			return nil, err
		}

		return &SearchKeyUnkeyword{Value: value}, nil

	case "unseen":
		return &SearchKeyUnseen{}, nil

	case "draft":
		return &SearchKeyDraft{}, nil

	case "header":
		field, err := parseStringKeyAString(p)
		if err != nil {
			return nil, err
		}

		value, err := parseStringKeyAString(p)
		if err != nil {
			return nil, err
		}

		return &SearchKeyHeader{Field: field, Value: value}, nil

	case "larger":
		value, err := parseStringKeyNumber(p)
		if err != nil {
			return nil, err
		}

		return &SearchKeyLarger{Value: value}, nil

	case "not":
		key, err := parseSearchKey(p)
		if err != nil {
			return nil, err
		}

		return &SearchKeyNot{Key: key}, nil

	case "or":
		key1, err := parseSearchKey(p)
		if err != nil {
			return nil, err
		}

		key2, err := parseSearchKey(p)
		if err != nil {
			return nil, err
		}

		return &SearchKeyOr{Key1: key1, Key2: key2}, nil

	case "sentbefore":
		value, err := parseStringKeyDate(p)
		if err != nil {
			return nil, err
		}

		return &SearchKeySentBefore{Value: value}, nil

	case "senton":
		value, err := parseStringKeyDate(p)
		if err != nil {
			return nil, err
		}

		return &SearchKeySentOn{Value: value}, nil

	case "sentsince":
		value, err := parseStringKeyDate(p)
		if err != nil {
			return nil, err
		}

		return &SearchKeySentSince{Value: value}, nil

	case "smaller":
		value, err := parseStringKeyNumber(p)
		if err != nil {
			return nil, err
		}

		return &SearchKeySmaller{Value: value}, nil

	case "uid":
		value, err := parseStringKeyNumber(p)
		if err != nil {
			return nil, err
		}

		if value < 0 || value > 0xFFFFFFFF {
			return nil, fmt.Errorf("invalid UID number")
		}

		return &SearchKeyUID{Value: uint32(value)}, nil

	case "undraft":
		return &SearchKeyUndraft{}, nil

	default:
		return nil, p.MakeErrorAtOffset(fmt.Sprintf("unknown search key '%v'", keyword.Value), keyword.Offset)
	}
}

func parseStringKeyAString(p *rfcparser.Parser) (string, error) {
	if err := p.Consume(rfcparser.TokenTypeSP, "expected space"); err != nil {
		return "", err
	}

	astring, err := p.ParseAString()
	if err != nil {
		return "", err
	}

	return astring.Value, nil
}

func parseStringKeyNumber(p *rfcparser.Parser) (int, error) {
	if err := p.Consume(rfcparser.TokenTypeSP, "expected space"); err != nil {
		return 0, err
	}

	return p.ParseNumber()
}

func parseStringKeyDate(p *rfcparser.Parser) (time.Time, error) {
	if err := p.Consume(rfcparser.TokenTypeSP, "expected space"); err != nil {
		return time.Time{}, err
	}

	return ParseDate(p)
}

func parseStringKeyAtom(p *rfcparser.Parser) (string, error) {
	if err := p.Consume(rfcparser.TokenTypeSP, "expected space"); err != nil {
		return "", err
	}

	return p.ParseAtom()
}
