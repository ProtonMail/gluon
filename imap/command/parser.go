package command

import (
	"fmt"
	rfcparser2 "github.com/ProtonMail/gluon/rfcparser"
	"strings"
)

type Builder interface {
	FromParser(p *rfcparser2.Parser) (Payload, error)
}

// Parser parses IMAP Commands.
type Parser struct {
	parser   *rfcparser2.Parser
	commands map[string]Builder
	lastTag  string
	lastCmd  string
}

func NewParser(s *rfcparser2.Scanner) *Parser {
	return NewParserWithLiteralContinuationCb(s, nil)
}

func NewParserWithLiteralContinuationCb(s *rfcparser2.Scanner, cb func() error) *Parser {
	return &Parser{
		parser: rfcparser2.NewParserWithLiteralContinuationCb(s, cb),
		commands: map[string]Builder{
			"list":        &ListCommandParser{},
			"append":      &AppendCommandParser{},
			"search":      &SearchCommandParser{},
			"fetch":       &FetchCommandParser{},
			"capability":  &CapabilityCommandParser{},
			"idle":        &IdleCommandParser{},
			"noop":        &NoopCommandParser{},
			"logout":      &LogoutCommandParser{},
			"check":       &CheckCommandParser{},
			"close":       &CloseCommandParser{},
			"expunge":     &ExpungeCommandParser{},
			"unselect":    &UnselectCommandParser{},
			"starttls":    &StartTLSCommandParser{},
			"status":      &StatusCommandParser{},
			"select":      &SelectCommandParser{},
			"examine":     &ExamineCommandParser{},
			"create":      &CreateCommandParser{},
			"delete":      &DeleteCommandParser{},
			"subscribe":   &SubscribeCommandParser{},
			"unsubscribe": &UnsubscribeCommandParser{},
			"rename":      &RenameCommandParser{},
			"lsub":        &LSubCommandParser{},
			"login":       &LoginCommandParser{},
			"store":       &StoreCommandParser{},
			"copy":        &CopyCommandParser{},
			"move":        &MoveCommandParser{},
			"uid":         NewUIDCommandParser(),
		},
	}
}

func (p *Parser) LastParsedTag() string {
	return p.lastTag
}

func (p *Parser) LastParsedCommand() string {
	return p.lastCmd
}

func (p *Parser) Parse() (Command, error) {
	result := Command{}

	p.lastTag = ""
	p.lastCmd = ""
	p.parser.ResetOffsetCounter()

	if err := p.parser.Advance(); err != nil {
		return result, err
	}

	tag, err := p.parseTag()
	if err != nil {
		return result, err
	}

	// Done command does not have a tag.
	if strings.ToLower(tag.Value) == "done" {
		p.lastCmd = "done"

		return Command{
			Tag:     "",
			Payload: &DoneCommand{},
		}, nil
	}

	result.Tag = tag.Value
	p.lastTag = tag.Value

	if err := p.parser.Consume(rfcparser2.TokenTypeSP, "Expected space after tag"); err != nil {
		return result, err
	}

	payload, err := p.parseCommand()
	if err != nil {
		return result, err
	}

	result.Payload = payload

	if err := p.parser.ConsumeNewLine(); err != nil {
		return result, err
	}

	return result, nil
}

func (p *Parser) parseCommand() (Payload, error) {
	var commandBytes []byte

	commandOffset := p.parser.CurrentToken().Offset

	for {
		if ok, err := p.parser.Matches(rfcparser2.TokenTypeChar); err != nil {
			return nil, err
		} else if ok {
			commandBytes = append(commandBytes, rfcparser2.ByteToLower(p.parser.PreviousToken().Value))
		} else {
			break
		}
	}

	p.lastCmd = string(commandBytes)

	builder, ok := p.commands[p.lastCmd]
	if !ok {
		return nil, p.parser.MakeErrorAtOffset(fmt.Sprintf("unknown command '%v'", p.lastCmd), commandOffset)
	}

	return builder.FromParser(p.parser)
}

func (p *Parser) parseTag() (rfcparser2.String, error) {
	// tag             = 1*<any ASTRING-CHAR except "+">
	isTagChar := func(tt rfcparser2.TokenType) bool {
		return rfcparser2.IsAStringChar(tt) && tt != rfcparser2.TokenTypePlus
	}

	if err := p.parser.ConsumeWith(isTagChar, "Invalid tag char detected"); err != nil {
		return rfcparser2.String{}, err
	}

	tag, err := p.parser.CollectBytesWhileMatchesWithPrevWith(isTagChar)
	if err != nil {
		return rfcparser2.String{}, err
	}

	return tag.IntoString(), err
}
