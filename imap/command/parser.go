package command

import (
	"fmt"
	rfcparser "github.com/ProtonMail/gluon/rfcparser"
	"strings"
)

type Builder interface {
	FromParser(p *rfcparser.Parser) (Payload, error)
}

// Parser parses IMAP Commands.
type Parser struct {
	parser   *rfcparser.Parser
	scanner  *rfcparser.Scanner
	commands map[string]Builder
	lastTag  string
	lastCmd  string
}

func NewParser(s *rfcparser.Scanner) *Parser {
	return NewParserWithLiteralContinuationCb(s, nil)
}

func NewParserWithLiteralContinuationCb(s *rfcparser.Scanner, cb func() error) *Parser {
	return &Parser{
		scanner: s,
		parser:  rfcparser.NewParserWithLiteralContinuationCb(s, cb),
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
			"id":          &IDCommandParser{},
		},
	}
}

func (p *Parser) LastParsedTag() string {
	return p.lastTag
}

func (p *Parser) LastParsedCommand() string {
	return p.lastCmd
}

// ConsumeInvalidInput will consume all remaining scanner input until a new line has been reached.
func (p *Parser) ConsumeInvalidInput() error {
	_, err := p.scanner.ConsumeUntilNewLine()

	return err
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
			Payload: &Done{},
		}, nil
	}

	result.Tag = tag.Value
	p.lastTag = tag.Value

	if err := p.parser.Consume(rfcparser.TokenTypeSP, "Expected space after tag"); err != nil {
		return result, err
	}

	payload, err := p.parseCommand()
	if err != nil {
		return result, err
	}

	result.Payload = payload

	if err := p.parser.Consume(rfcparser.TokenTypeCR, "expected CR"); err != nil {
		return Command{}, err
	}

	// Can't fully consume the last new line here or we will hang forever as the clients don't send the next token.
	// In the next loop, the call to advance will ensure the next token in the stream gets loaded properly.
	if !p.parser.Check(rfcparser.TokenTypeLF) {
		return Command{}, p.parser.MakeError("expected LF after CR")
	}

	return result, nil
}

func (p *Parser) parseCommand() (Payload, error) {
	var commandBytes []byte

	commandOffset := p.parser.CurrentToken().Offset

	for {
		if ok, err := p.parser.Matches(rfcparser.TokenTypeChar); err != nil {
			return nil, err
		} else if ok {
			commandBytes = append(commandBytes, rfcparser.ByteToLower(p.parser.PreviousToken().Value))
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

func (p *Parser) parseTag() (rfcparser.String, error) {
	// tag             = 1*<any ASTRING-CHAR except "+">
	isTagChar := func(tt rfcparser.TokenType) bool {
		return rfcparser.IsAStringChar(tt) && tt != rfcparser.TokenTypePlus
	}

	if err := p.parser.ConsumeWith(isTagChar, "Invalid tag char detected"); err != nil {
		return rfcparser.String{}, err
	}

	tag, err := p.parser.CollectBytesWhileMatchesWithPrevWith(isTagChar)
	if err != nil {
		return rfcparser.String{}, err
	}

	return tag.IntoString(), err
}
