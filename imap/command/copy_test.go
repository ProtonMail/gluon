package command

import (
	"bytes"
	"github.com/ProtonMail/gluon/imap/parser"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParser_CopyCommand(t *testing.T) {
	input := toIMAPLine(`tag COPY 1:* INBOX`)
	s := parser.NewScanner(bytes.NewReader(input))
	p := NewParser(s)

	expected := Command{Tag: "tag", Payload: &CopyCommand{
		Mailbox: "INBOX",
		SeqSet:  []SeqRange{{Begin: 1, End: SeqNumValueAsterisk}},
	}}

	cmd, err := p.Parse()
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
	require.Equal(t, "copy", p.LastParsedCommand())
	require.Equal(t, "tag", p.LastParsedTag())
}
