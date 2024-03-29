package command

import (
	"bytes"
	"testing"

	"github.com/ProtonMail/gluon/rfcparser"
	"github.com/stretchr/testify/require"
)

func TestParser_MoveCommand(t *testing.T) {
	input := toIMAPLine(`tag MOVE 1:* INBOX`)
	s := rfcparser.NewScanner(bytes.NewReader(input))
	p := NewParser(s)

	expected := Command{Tag: "tag", Payload: &Move{
		Mailbox: "INBOX",
		SeqSet:  []SeqRange{{Begin: 1, End: SeqNumValueAsterisk}},
	}}

	cmd, err := p.Parse()
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
	require.Equal(t, "move", p.LastParsedCommand())
	require.Equal(t, "tag", p.LastParsedTag())
}
