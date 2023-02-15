package command

import (
	"bytes"
	rfcparser "github.com/ProtonMail/gluon/rfcparser"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestDateTimeParser(t *testing.T) {
	input := []byte(`"15-Nov-1984 13:37:01 +0730"`)
	expected := buildAppendDateTime(1984, time.November, 15, 13, 37, 1, 07, 30, false)

	p := rfcparser.NewParser(rfcparser.NewScanner(bytes.NewReader(input)))
	// Advance at least once to prepare first token.
	err := p.Advance()
	require.NoError(t, err)

	dt, err := ParseDateTime(p)
	require.NoError(t, err)
	require.Equal(t, expected, dt)
}

func TestDateTimeParser_OneDayDigit(t *testing.T) {
	input := []byte(`" 5-Nov-1984 13:37:01 -0730"`)
	expected := buildAppendDateTime(1984, time.November, 5, 13, 37, 1, 07, 30, true)

	p := rfcparser.NewParser(rfcparser.NewScanner(bytes.NewReader(input)))
	// Advance at least once to prepare first token.
	err := p.Advance()
	require.NoError(t, err)

	dt, err := ParseDateTime(p)
	require.NoError(t, err)
	require.Equal(t, expected, dt)
}
