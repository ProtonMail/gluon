package command

import (
	"bytes"
	"testing"

	"github.com/ProtonMail/gluon/rfcparser"
	"github.com/stretchr/testify/require"
)

func TestParser_DeleteCommand(t *testing.T) {
	input := toIMAPLine(`tag DELETE INBOX`)
	s := rfcparser.NewScanner(bytes.NewReader(input))
	p := NewParser(s)

	expected := Command{Tag: "tag", Payload: &Delete{
		Mailbox: "INBOX",
	}}

	cmd, err := p.Parse()
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
	require.Equal(t, "delete", p.LastParsedCommand())
	require.Equal(t, "tag", p.LastParsedTag())
}
