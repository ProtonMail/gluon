package command

import (
	"bytes"
	"github.com/ProtonMail/gluon/rfcparser"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParser_UnsubscribeCommand(t *testing.T) {
	input := toIMAPLine(`tag UNSUBSCRIBE INBOX`)
	s := rfcparser.NewScanner(bytes.NewReader(input))
	p := NewParser(s)

	expected := Command{Tag: "tag", Payload: &Unsubscribe{
		Mailbox: "INBOX",
	}}

	cmd, err := p.Parse()
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
	require.Equal(t, "unsubscribe", p.LastParsedCommand())
	require.Equal(t, "tag", p.LastParsedTag())
}
