package parser

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"testing"
)

func newTestParser(input []byte) *Parser {
	p := NewParser(NewScanner(bytes.NewReader(input)))
	if err := p.Advance(); err != nil {
		panic("Could not advance parser to first token")
	}

	return p
}

func TestParser_ParseNumber(t *testing.T) {
	input := []byte(`1024`)
	p := newTestParser(input)

	v, err := p.ParseNumber()
	require.NoError(t, err)
	require.Equal(t, 1024, v)
}

func TestParser_ParseNumberInvalid(t *testing.T) {
	inputs := [][]byte{
		[]byte(`-1`),
		[]byte(`.1`),
		[]byte(`a`),
		[]byte(`+1`),
	}
	for _, i := range inputs {
		p := newTestParser(i)

		_, err := p.ParseNumber()
		require.Error(t, err)
	}
}

func TestParser_ParseQuoted(t *testing.T) {
	values := map[string]string{
		`"hello world 10234"`:  `hello world 10234`,
		`"h\"b\"c"`:            `h"b"c`,
		`"\\{}@foo&^*(#$!<>="`: `\{}@foo&^*(#$!<>=`,
	}

	for input, expected := range values {
		p := newTestParser([]byte(input))
		v, err := p.ParseQuoted()
		require.NoError(t, err)
		require.Equal(t, expected, v)
	}
}

func TestParser_ParseQuotedInvalid(t *testing.T) {
	inputs := [][]byte{
		[]byte(`"foo`),
		[]byte(`"x\f"`),
		[]byte(`"\/"'`),
		[]byte(`foo"'`),
	}
	for _, i := range inputs {
		p := newTestParser(i)

		_, err := p.ParseNumber()
		require.Error(t, err)
	}
}

func TestParser_ParseString(t *testing.T) {
	// Strings are either quoted or literals.
	values := map[string]string{
		`"hello world 10234"`: `hello world 10234`,
		"{5}\r\n01234":        `01234`,
	}

	for input, expected := range values {
		p := newTestParser([]byte(input))
		v, err := p.ParseString()
		require.NoError(t, err)
		require.Equal(t, expected, v)
	}
}

func TestParser_ParseLiteral(t *testing.T) {
	// Strings are either quoted or literals.
	values := map[string]string{
		"{5}\r\n h123": ` h123`,
		"{6}\r\n你好":    `你好`,
	}

	for input, expected := range values {
		p := newTestParser([]byte(input))
		v, err := p.ParseLiteral()
		require.NoError(t, err)
		require.Equal(t, []byte(expected), v)
	}
}

func TestParser_ParseAString(t *testing.T) {
	values := map[string]string{
		"{5}\r\n h123":         ` h123`,
		"{6}\r\n你好":            `你好`,
		`"hello world 10234"`:  `hello world 10234`,
		`"h\"b\"c"`:            `h"b"c`,
		`"\\{}@foo&^*(#$!<>="`: `\{}@foo&^*(#$!<>=`,
		`hello_world`:          `hello_world`,
		`hello-1234`:           `hello-1234`,
	}

	for input, expected := range values {
		p := newTestParser([]byte(input))
		v, err := p.ParseAString()
		require.NoError(t, err)
		require.Equal(t, expected, v)
	}
}

func TestParser_ParseFlagList(t *testing.T) {
	values := map[string][]string{
		`(\Answered)`:                {`\Answered`},
		`(\Answered Foo \Something)`: {`\Answered`, `Foo`, `\Something`},
	}

	for input, expected := range values {
		p := newTestParser([]byte(input))
		v, err := p.ParseFlagList()
		require.NoError(t, err)
		require.Equal(t, expected, v)
	}
}

func TestParser_ParseFlagListInvalid(t *testing.T) {
	inputs := [][]byte{
		[]byte(`()`),
		[]byte(`(\Foo\Bar)`),
		[]byte(`"(\Recent)`),
		[]byte(`(\Foo )`),
		[]byte(`(\Foo`),
	}
	for _, i := range inputs {
		p := newTestParser(i)

		_, err := p.ParseNumber()
		require.Error(t, err)
	}
}
