package command

import (
	"bytes"
	"testing"

	"github.com/ProtonMail/gluon/rfcparser"
	"github.com/stretchr/testify/require"
)

func TestParser_StartTLSCommand(t *testing.T) {
	input := toIMAPLine(`tag STARTTLS`)
	s := rfcparser.NewScanner(bytes.NewReader(input))
	p := NewParser(s)

	expected := Command{Tag: "tag", Payload: &StartTLS{}}

	cmd, err := p.Parse()
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
	require.Equal(t, "starttls", p.LastParsedCommand())
	require.Equal(t, "tag", p.LastParsedTag())
}
