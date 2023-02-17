package rfc5322

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseFWS(t *testing.T) {
	inputs := []string{
		" \t ",
		"\r\n\t",
		" \r\n\t",
		"  \r\n  \r\n  \r\n\t",
		" \t\r\n    ",
	}

	for _, i := range inputs {
		p := newTestRFCParser(i)
		err := parseFWS(p)
		require.NoError(t, err)
	}
}

func TestParserComment(t *testing.T) {
	inputs := []string{
		"(my comment here)",
		"(my comment here )",
		"( my comment here)",
		"( my comment here )",
		"(my\r\n comment here)",
		"(my\r\n (comment) here)",
		"(\\my\r\n (comment) here)",
		"(" + string([]byte{0x7F, 0x8}) + ")",
	}

	for _, i := range inputs {
		p := newTestRFCParser(i)
		err := parseComment(p)
		require.NoError(t, err)
	}
}

func TestParserCFWS(t *testing.T) {
	inputs := []string{
		" ",
		"(my comment here)",
		" (my comment here) ",
		" \r\n (my comment here)  ",
		" \r\n \r\n (my comment here) \r\n ",
	}

	for _, i := range inputs {
		p := newTestRFCParser(i)
		err := parseCFWS(p)
		require.NoError(t, err)
	}
}
