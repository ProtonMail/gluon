package command

import (
	"bytes"
	"testing"

	"github.com/ProtonMail/gluon/rfcparser"
	"github.com/stretchr/testify/require"
)

func TestParser_SubscribeCommand(t *testing.T) {
	input := toIMAPLine(`tag SUBSCRIBE INBOX`)
	s := rfcparser.NewScanner(bytes.NewReader(input))
	p := NewParser(s)

	expected := Command{Tag: "tag", Payload: &Subscribe{
		Mailbox: "INBOX",
	}}

	cmd, err := p.Parse()
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
	require.Equal(t, "subscribe", p.LastParsedCommand())
	require.Equal(t, "tag", p.LastParsedTag())
}
