package command

import (
	"bytes"
	"github.com/ProtonMail/gluon/rfcparser"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParser_IDCommandGet(t *testing.T) {
	input := toIMAPLine(`tag ID NIL`)
	s := rfcparser.NewScanner(bytes.NewReader(input))
	p := NewParser(s)

	expected := Command{Tag: "tag", Payload: &IDGet{}}

	cmd, err := p.Parse()
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
	require.Equal(t, "id", p.LastParsedCommand())
	require.Equal(t, "tag", p.LastParsedTag())
}

func TestParser_IDCommandSetOne(t *testing.T) {
	input := toIMAPLine(`tag ID ("foo" "bar")`)
	s := rfcparser.NewScanner(bytes.NewReader(input))
	p := NewParser(s)

	expected := Command{Tag: "tag", Payload: &IDSet{Values: map[string]string{"foo": "bar"}}}

	cmd, err := p.Parse()
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
	require.Equal(t, "id", p.LastParsedCommand())
	require.Equal(t, "tag", p.LastParsedTag())
}

func TestParser_IDCommandSetEmpty(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &IDSet{
		Values: map[string]string{},
	}}

	cmd, err := testParseCommand(`tag ID ()`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_IDCommandSetMany(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &IDSet{
		Values: map[string]string{
			"foo": "bar",
			"a":   "",
			"c":   "d",
		},
	}}

	cmd, err := testParseCommand(`tag ID ("foo" "bar" "a" NIL "c" "d")`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_IDCommandFailures(t *testing.T) {
	inputs := []string{
		"tag ID",
		"tag ID ",
		"tag ID N",
		"tag ID (",
		`tag ID ("foo")`,
		`tag ID ("foo" )`,
		`tag ID ("foo""bar")`,
		`tag ID (nil nil)`,
		`tag ID ("foo" "bar"`,
		`tag ID ("foo" "bar" "z")`,
	}

	for _, i := range inputs {
		_, err := testParseCommand(i)
		require.Error(t, err)
	}
}
