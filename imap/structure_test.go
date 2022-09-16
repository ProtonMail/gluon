package imap

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmbeddedRFC822WithoutHeader(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "rfc822.eml"))
	require.NoError(t, err)

	parsed, err := NewParsedMessage(b)
	require.NoError(t, err)
	require.NotNil(t, parsed)
}

func TestHeaderOutOfBounds(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "bounds.eml"))
	require.NoError(t, err)

	parsed, err := NewParsedMessage(b)
	require.NoError(t, err)
	require.NotNil(t, parsed)
}
