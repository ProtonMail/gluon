package command

import (
	"fmt"
	"github.com/ProtonMail/gluon/rfcparser"
)

type List struct {
	Mailbox     string
	ListMailbox string
}

func (l List) String() string {
	return fmt.Sprintf("LIST '%v' '%v'", l.Mailbox, l.ListMailbox)
}

func (l List) SanitizedString() string {
	return l.String()
}

type ListCommandParser struct{}

func (ListCommandParser) FromParser(p *rfcparser.Parser) (Payload, error) {
	// list            = "LIST" SP mailbox SP list-mailbox
	if err := p.Consume(rfcparser.TokenTypeSP, "expected space after command"); err != nil {
		return nil, err
	}

	mailbox, err := ParseMailbox(p)
	if err != nil {
		return nil, err
	}

	if err := p.Consume(rfcparser.TokenTypeSP, "expected space after mailbox"); err != nil {
		return nil, err
	}

	listMailbox, err := parseListMailbox(p)
	if err != nil {
		return nil, err
	}

	return &List{
		Mailbox:     mailbox.Value,
		ListMailbox: listMailbox.Value,
	}, nil
}

func parseListMailbox(p *rfcparser.Parser) (rfcparser.String, error) {
	/*
	  list-mailbox    = 1*list-char / string

	  list-char       = ATOM-CHAR / list-wildcards / resp-specials

	  list-wildcards  = "%" / "*"
	*/
	isListChar := func(tt rfcparser.TokenType) bool {
		return rfcparser.IsAtomChar(tt) || rfcparser.IsRespSpecial(tt) || tt == rfcparser.TokenTypePercent || tt == rfcparser.TokenTypeAsterisk
	}

	if ok, err := p.MatchesWith(isListChar); err != nil {
		return rfcparser.String{}, err
	} else if !ok {
		return p.ParseString()
	}

	listMailbox, err := p.CollectBytesWhileMatchesWithPrevWith(isListChar)
	if err != nil {
		return rfcparser.String{}, err
	}

	return listMailbox.IntoString(), nil
}
