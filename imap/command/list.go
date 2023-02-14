package command

import (
	"fmt"
	rfcparser2 "github.com/ProtonMail/gluon/rfcparser"
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

func (ListCommandParser) FromParser(p *rfcparser2.Parser) (Payload, error) {
	// list            = "LIST" SP mailbox SP list-mailbox
	if err := p.Consume(rfcparser2.TokenTypeSP, "expected space after command"); err != nil {
		return nil, err
	}

	mailbox, err := p.ParseMailbox()
	if err != nil {
		return nil, err
	}

	if err := p.Consume(rfcparser2.TokenTypeSP, "expected space after mailbox"); err != nil {
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

func parseListMailbox(p *rfcparser2.Parser) (string, error) {
	/*
	  list-mailbox    = 1*list-char / string

	  list-char       = ATOM-CHAR / list-wildcards / resp-specials

	  list-wildcards  = "%" / "*"
	*/
	isListChar := func(tt rfcparser2.TokenType) bool {
		return rfcparser2.IsAtomChar(tt) || rfcparser2.IsRespSpecial(tt) || tt == rfcparser2.TokenTypePercent || tt == rfcparser2.TokenTypeAsterisk
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
