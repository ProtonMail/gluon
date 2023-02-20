package command

import (
	"bytes"
	"testing"

	"github.com/ProtonMail/gluon/rfcparser"
	"github.com/stretchr/testify/require"
)

func TestParser_IdleCommand(t *testing.T) {
	input := toIMAPLine(`tag IDLE`)
	s := rfcparser.NewScanner(bytes.NewReader(input))
	p := NewParser(s)

	expected := Command{Tag: "tag", Payload: &Idle{}}

	cmd, err := p.Parse()
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
	require.Equal(t, "idle", p.LastParsedCommand())
	require.Equal(t, "tag", p.LastParsedTag())
}
