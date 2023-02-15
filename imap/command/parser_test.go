package command

import (
	"bytes"
	"fmt"
	rfcparser "github.com/ProtonMail/gluon/rfcparser"
	"github.com/stretchr/testify/require"
	"io"
	"testing"
	"time"
)

func toIMAPLine(string ...string) []byte {
	var result []byte

	for _, v := range string {
		result = append(result, []byte(v)...)
		result = append(result, '\r', '\n')
	}

	return result
}

func testParseCommand(lines ...string) (Command, error) {
	input := toIMAPLine(lines...)
	s := rfcparser.NewScanner(bytes.NewReader(input))
	p := NewParser(s)

	return p.Parse()
}

func TestParser_InvalidTag(t *testing.T) {
	input := []byte(`+tag LIST "" "*"`)
	s := rfcparser.NewScanner(bytes.NewReader(input))
	p := NewParser(s)

	_, err := p.Parse()
	require.Error(t, err)
	require.Empty(t, p.LastParsedCommand())
	require.Empty(t, p.LastParsedTag())
}

func TestParser_TestEof(t *testing.T) {
	var input []byte
	s := rfcparser.NewScanner(bytes.NewReader(input))
	p := NewParser(s)

	_, err := p.Parse()
	require.Error(t, err)
	require.True(t, rfcparser.IsError(err))
	parserError, ok := err.(*rfcparser.Error) //nolint:errorlint
	require.True(t, ok)
	require.True(t, parserError.IsEOF())
}

func TestParser_InvalidFollowedByValidCommand(t *testing.T) {
	input := toIMAPLine(`+tag LIST "" "*"`, `foo list "bar" "*"`)
	s := rfcparser.NewScanner(bytes.NewReader(input))
	p := NewParser(s)

	// First command fails.
	_, err := p.Parse()
	require.Error(t, err)
	require.True(t, rfcparser.IsError(err))

	// Clear any other input until new line has been reached
	err = p.ConsumeInvalidInput()
	require.NoError(t, err)

	// Second command should succeed.
	expected := Command{Tag: "foo", Payload: &List{
		Mailbox:     "bar",
		ListMailbox: "*",
	}}

	cmd, err := p.Parse()
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_LiteralWithContinuationSubmission(t *testing.T) {
	// Run one go routine that submits bytes until the continuation call has been received
	continueCh := make(chan struct{})
	reader, writer := io.Pipe()

	go func() {
		defer writer.Close()

		firstLine := toIMAPLine(`A003 APPEND saved-messages (\Seen) "15-Nov-1984 13:37:01 +0730" {23}`)

		secondLine := toIMAPLine(`My message body is here`)

		if l, err := writer.Write(firstLine); err != nil || l != len(firstLine) {
			writer.CloseWithError(fmt.Errorf("failed to write first line: %w", err))
			return
		}

		<-continueCh

		if l, err := writer.Write(secondLine); err != nil || l != len(secondLine) {
			writer.CloseWithError(fmt.Errorf("failed to write second line: %w", err))
			return
		}
	}()

	s := rfcparser.NewScanner(reader)
	p := NewParserWithLiteralContinuationCb(s, func() error {
		close(continueCh)
		return nil
	})

	expected := Command{Tag: "A003", Payload: &Append{
		Mailbox:  "saved-messages",
		Flags:    []string{`\Seen`},
		Literal:  []byte("My message body is here"),
		DateTime: buildAppendDateTime(1984, time.November, 15, 13, 37, 1, 07, 30, false),
	}}

	cmd, err := p.Parse()
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}
