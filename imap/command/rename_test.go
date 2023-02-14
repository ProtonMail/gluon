package command

import (
	"bytes"
	"github.com/ProtonMail/gluon/rfcparser"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParser_RenameCommand(t *testing.T) {
	input := toIMAPLine(`tag RENAME Foo Bar`)
	s := rfcparser.NewScanner(bytes.NewReader(input))
	p := NewParser(s)

	expected := Command{Tag: "tag", Payload: &RenameCommand{
		From: "Foo",
		To:   "Bar",
	}}

	cmd, err := p.Parse()
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
	require.Equal(t, "rename", p.LastParsedCommand())
	require.Equal(t, "tag", p.LastParsedTag())
}
