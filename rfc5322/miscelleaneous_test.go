package rfc5322

import (
	"testing"

	"github.com/bradenaw/juniper/xslices"
	"github.com/stretchr/testify/require"
)

func TestParseWord(t *testing.T) {
	inputs := map[string]string{
		`"f\".c"`:                "f\".c",
		"\" \r\n f\\\".c\r\n \"": " f\".c ",
		` " foo bar derer " `:    " foo bar derer ",
		`foo`:                    "foo",
	}

	for i, e := range inputs {
		p := newTestRFCParser(i)
		v, err := parseWord(p)
		require.NoError(t, err)
		require.Equal(t, e, v.String.Value)
	}
}

func TestParsePhrase(t *testing.T) {
	inputs := map[string][]string{
		`foo "quoted"`:     {"foo", "quoted"},
		`"f\".c" "quoted"`: {"f\".c", "quoted"},
		`foo bar`:          {"foo", "bar"},
		`foo.bar`:          {"foo", ".", "bar"},
		`foo . bar`:        {"foo", ".", "bar"},
	}

	for i, e := range inputs {
		p := newTestRFCParser(i)
		v, err := parsePhrase(p)
		require.NoError(t, err)
		require.Equal(t, e, xslices.Map(v, func(v parserString) string { return v.String.Value }))
	}
}
