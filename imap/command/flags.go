package command

import (
	"fmt"
	"github.com/ProtonMail/gluon/imap/parser"
	"strings"
)

func TryParseFlagList(p *parser.Parser) ([]string, bool, error) {
	if !p.Check(parser.TokenTypeLParen) {
		return nil, false, nil
	}

	flags, err := ParseFlagList(p)

	return flags, true, err
}

func ParseFlagList(p *parser.Parser) ([]string, error) {
	// flag-list       = "(" [flag *(SP flag)] ")"
	var flags []string

	if err := p.Consume(parser.TokenTypeLParen, "Expected '(' at start of flag list"); err != nil {
		return nil, err
	}

	{
		firstFlag, err := ParseFlag(p)
		if err != nil {
			return nil, err
		}
		flags = append(flags, firstFlag)
	}

	for {
		if ok, err := p.Matches(parser.TokenTypeSP); err != nil {
			return nil, err
		} else if !ok {
			break
		}

		flag, err := ParseFlag(p)
		if err != nil {
			return nil, err
		}

		flags = append(flags, flag)
	}

	if err := p.Consume(parser.TokenTypeRParen, "Expected ')' at end of flag list"); err != nil {
		return nil, err
	}

	return flags, nil
}

func ParseFlag(p *parser.Parser) (string, error) {
	/*
	 flag            = "\Answered" / "\Flagged" / "\Deleted" /
	                   "\Seen" / "\Draft" / flag-keyword / flag-extension
	                     ; Does not include "\Recent"

	 flag-extension  = "\" atom
	                     ; Future expansion.  Client implementations
	                     ; MUST accept flag-extension flags.  Server
	                     ; implementations MUST NOT generate
	                     ; flag-extension flags except as defined by
	                     ; future standard or standards-track
	                     ; revisions of this specification.

	 flag-keyword    = atom
	*/
	hasBackslash, err := p.Matches(parser.TokenTypeBackslash)
	if err != nil {
		return "", err
	}

	if hasBackslash {
		flag, err := p.ParseAtom()
		if err != nil {
			return "", err
		}

		if strings.EqualFold(flag, "recent") {
			return "", p.MakeError("Recent Flag is not allowed in this context")
		}

		return fmt.Sprintf("\\%v", flag), nil
	}

	return p.ParseAtom()
}
