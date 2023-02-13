package command

import (
	"fmt"
	"github.com/ProtonMail/gluon/imap/parser"
)

type ListCommand struct {
	Mailbox     string
	ListMailbox string
}

func (l ListCommand) String() string {
	return fmt.Sprintf("LIST '%v' '%v'", l.Mailbox, l.ListMailbox)
}

func (l ListCommand) SanitizedString() string {
	return l.String()
}

type ListCommandParser struct{}

func (ListCommandParser) FromParser(p *parser.Parser) (Payload, error) {
	// list            = "LIST" SP mailbox SP list-mailbox
	if err := p.Consume(parser.TokenTypeSP, "expected space after command"); err != nil {
		return nil, err
	}

	mailbox, err := p.ParseMailbox()
	if err != nil {
		return nil, err
	}

	if err := p.Consume(parser.TokenTypeSP, "expected space after mailbox"); err != nil {
		return nil, err
	}

	listMailbox, err := parseListMailbox(p)
	if err != nil {
		return nil, err
	}

	return &ListCommand{
		Mailbox:     mailbox,
		ListMailbox: listMailbox,
	}, nil
}

func parseListMailbox(p *parser.Parser) (string, error) {
	/*
	  list-mailbox    = 1*list-char / string

	  list-char       = ATOM-CHAR / list-wildcards / resp-specials

	  list-wildcards  = "%" / "*"
	*/
	isListChar := func(tt parser.TokenType) bool {
		return parser.IsAtomChar(tt) || parser.IsRespSpecial(tt) || tt == parser.TokenTypePercent || tt == parser.TokenTypeAsterisk
	}

	if ok, err := p.MatchesWith(isListChar); err != nil {
		return "", err
	} else if !ok {
		return p.ParseString()
	}

	listMailbox, err := p.CollectBytesWhileMatchesWithPrevWith(isListChar)
	if err != nil {
		return "", err
	}

	return string(listMailbox), nil
}
