package command

import (
	"bytes"
	rfcparser2 "github.com/ProtonMail/gluon/rfcparser"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseSeqSet(t *testing.T) {
	input := []byte(`1:*,*,20,40:30`)
	expected := []SeqRange{
		{
			Begin: SeqNum(1),
			End:   SeqNumValueAsterisk,
		},
		{
			Begin: SeqNumValueAsterisk,
			End:   SeqNumValueAsterisk,
		},
		{
			Begin: SeqNum(20),
			End:   SeqNum(20),
		},
		{
			Begin: SeqNum(40),
			End:   SeqNum(30),
		},
	}

	p := rfcparser2.NewParser(rfcparser2.NewScanner(bytes.NewReader(input)))
	// Advance at least once to prepare first token.
	err := p.Advance()
	require.NoError(t, err)

	dt, err := ParseSeqSet(p)
	require.NoError(t, err)
	require.Equal(t, expected, dt)
}
