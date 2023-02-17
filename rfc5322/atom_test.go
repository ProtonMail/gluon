package rfc5322

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseDotAtom(t *testing.T) {
	inputs := map[string]string{
		"foobar.!#$%'*+-=?^~_{}`|/": "foobar.!#$%'*+-=?^~_{}`|/",
		"  f.b  ":                   "f.b",
		" \r\n f.b":                 "f.b",
		" \r\n f.b \r\n ":           "f.b",
	}

	for i, e := range inputs {
		p := newTestRFCParser(i)
		v, err := parseDotAtom(p)
		require.NoError(t, err)
		require.Equal(t, e, v.Value)
	}
}

func TestParseAtom(t *testing.T) {
	inputs := map[string]string{
		"foobar!#$%'*+-=?^~_{}`|/": "foobar!#$%'*+-=?^~_{}`|/",
		"  fb  ":                   "fb",
		" \r\n fb":                 "fb",
		" \r\n fb \r\n ":           "fb",
	}

	for i, e := range inputs {
		p := newTestRFCParser(i)
		v, err := parseDotAtom(p)
		require.NoError(t, err)
		require.Equal(t, e, v.Value)
	}
}
