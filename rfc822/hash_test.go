package rfc822

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestGetMessageHashSameBodyDifferentTextEncodings(t *testing.T) {
	data1, err := os.ReadFile("testdata/hash_quoted.eml")
	require.NoError(t, err)

	data2, err := os.ReadFile("testdata/hash_utf8.eml")
	require.NoError(t, err)

	h1, err := GetMessageHash(data1)
	require.NoError(t, err)

	h2, err := GetMessageHash(data2)
	require.NoError(t, err)

	require.Equal(t, h1, h2)
}
