package rfc5322

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQuotedString(t *testing.T) {
	inputs := map[string]string{
		`"f\".c"`:                "f\".c",
		"\" \r\n f\\\".c\r\n \"": " f\".c ",
		` " foo bar derer " `:    " foo bar derer ",
	}

	for i, e := range inputs {
		p := newTestRFCParser(i)
		v, err := parseQuotedString(p)
		require.NoError(t, err)
		require.Equal(t, e, v.String.Value)
	}
}
