package command

import (
	"bytes"
	"testing"

	cppParser "github.com/ProtonMail/gluon/internal/parser"
	"github.com/ProtonMail/gluon/rfcparser"
	"github.com/stretchr/testify/require"
)

func TestParser_ListCommandQuoted(t *testing.T) {
	input := toIMAPLine(`tag LIST "" "*"`)
	s := rfcparser.NewScanner(bytes.NewReader(input))
	p := NewParser(s)

	expected := Command{Tag: "tag", Payload: &List{
		Mailbox:     "",
		ListMailbox: "*",
	}}

	cmd, err := p.Parse()
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
	require.Equal(t, "list", p.LastParsedCommand())
	require.Equal(t, "tag", p.LastParsedTag())
}

func TestParser_ListCommandSpecialAsterisk(t *testing.T) {
	input := toIMAPLine(`tag LIST "foo" *`)
	s := rfcparser.NewScanner(bytes.NewReader(input))
	p := NewParser(s)

	expected := Command{Tag: "tag", Payload: &List{
		Mailbox:     "foo",
		ListMailbox: "*",
	}}

	cmd, err := p.Parse()
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
	require.Equal(t, "list", p.LastParsedCommand())
	require.Equal(t, "tag", p.LastParsedTag())
}

func TestParser_ListCommandSpecialPercentage(t *testing.T) {
	input := toIMAPLine(`tag LIST "bar" %`)
	s := rfcparser.NewScanner(bytes.NewReader(input))
	p := NewParser(s)

	expected := Command{Tag: "tag", Payload: &List{
		Mailbox:     "bar",
		ListMailbox: "%",
	}}

	cmd, err := p.Parse()
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
	require.Equal(t, "list", p.LastParsedCommand())
	require.Equal(t, "tag", p.LastParsedTag())
}

func TestParser_ListCommandLiteral(t *testing.T) {
	input := toIMAPLine(`tag LIST {5}`, `"bar" %`)
	s := rfcparser.NewScanner(bytes.NewReader(input))
	continuationCalled := false
	p := NewParserWithLiteralContinuationCb(s, func() error {
		continuationCalled = true
		return nil
	})
	expected := Command{Tag: "tag", Payload: &List{
		Mailbox:     `"bar"`,
		ListMailbox: "%",
	}}

	cmd, err := p.Parse()
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
	require.True(t, continuationCalled)
	require.Equal(t, "list", p.LastParsedCommand())
	require.Equal(t, "tag", p.LastParsedTag())
}

func BenchmarkParser_ListCommand(b *testing.B) {
	input := toIMAPLine(`tag LIST "bar" %`)

	for i := 0; i < b.N; i++ {
		s := rfcparser.NewScanner(bytes.NewReader(input))
		p := NewParser(s)

		_, err := p.Parse()
		require.NoError(b, err)
	}
}

func BenchmarkParser_ListCommandCPP(b *testing.B) {
	input := string(toIMAPLine(`tag LIST "bar" %`))
	literalMap := cppParser.NewStringMap()

	for i := 0; i < b.N; i++ {
		cppParser.Parse(input, literalMap, "/")
	}
}
