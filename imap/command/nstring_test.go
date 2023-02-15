package command

import (
	"bytes"
	"github.com/ProtonMail/gluon/rfcparser"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseNStringString(t *testing.T) {
	input := []byte(`"foo"`)

	p := rfcparser.NewParser(rfcparser.NewScanner(bytes.NewReader(input)))
	// Advance at least once to prepare first token.
	err := p.Advance()
	require.NoError(t, err)

	v, isNil, err := ParseNString(p)
	require.NoError(t, err)
	require.Equal(t, "foo", v.Value)
	require.False(t, isNil)
}

func TestParseNStringNIL(t *testing.T) {
	input := []byte(`NIL`)

	p := rfcparser.NewParser(rfcparser.NewScanner(bytes.NewReader(input)))
	// Advance at least once to prepare first token.
	err := p.Advance()
	require.NoError(t, err)

	_, isNil, err := ParseNString(p)
	require.NoError(t, err)
	require.True(t, isNil)
}
