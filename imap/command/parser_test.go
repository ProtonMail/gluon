package command

import (
	"bytes"
	rfcparser2 "github.com/ProtonMail/gluon/rfcparser"
	"github.com/stretchr/testify/require"
	"testing"
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
	s := rfcparser2.NewScanner(bytes.NewReader(input))
	p := NewParser(s)

	return p.Parse()
}

func TestParser_InvalidTag(t *testing.T) {
	input := []byte(`+tag LIST "" "*"`)
	s := rfcparser2.NewScanner(bytes.NewReader(input))
	p := NewParser(s)

	_, err := p.Parse()
	require.Error(t, err)
	require.Empty(t, p.LastParsedCommand())
	require.Empty(t, p.LastParsedTag())
}

func TestParser_TestEof(t *testing.T) {
	var input []byte
	s := rfcparser2.NewScanner(bytes.NewReader(input))
	p := NewParser(s)

	_, err := p.Parse()
	require.Error(t, err)
	require.True(t, rfcparser2.IsError(err))
	parserError, ok := err.(*rfcparser2.Error) //nolint:errorlint
	require.True(t, ok)
	require.True(t, parserError.IsEOF())
}
