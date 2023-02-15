package command

import (
	"bufio"
	"bytes"
	"github.com/ProtonMail/gluon/rfcparser"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestInputCollector(t *testing.T) {
	input := toIMAPLine(`A003 APPEND saved-messages (\Seen) "15-Nov-1984 13:37:01 +0730" {23}`, `My message body is here`)
	source := bufio.NewReader(bytes.NewReader(input))
	collector := NewInputCollector(source)

	s := rfcparser.NewScannerWithReader(collector)
	p := NewParser(s)

	expected := Command{Tag: "A003", Payload: &Append{
		Mailbox:  "saved-messages",
		Flags:    []string{`\Seen`},
		Literal:  []byte("My message body is here"),
		DateTime: buildAppendDateTime(1984, time.November, 15, 13, 37, 1, 07, 30, false),
	}}

	cmd, err := p.Parse()
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
	require.Equal(t, "append", p.LastParsedCommand())
	require.Equal(t, "A003", p.LastParsedTag())
	require.Equal(t, input, collector.Bytes())
}