package command

import (
	"bytes"
	"github.com/ProtonMail/gluon/rfcparser"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParser_ExpungeCommand(t *testing.T) {
	input := toIMAPLine(`tag EXPUNGE`)
	s := rfcparser.NewScanner(bytes.NewReader(input))
	p := NewParser(s)

	expected := Command{Tag: "tag", Payload: &Expunge{}}

	cmd, err := p.Parse()
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
	require.Equal(t, "expunge", p.LastParsedCommand())
	require.Equal(t, "tag", p.LastParsedTag())
}
