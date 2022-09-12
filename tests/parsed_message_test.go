package tests

import (
	"github.com/ProtonMail/gluon/imap"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkParsedMessage(b *testing.B) {
	literal, err := os.ReadFile("testdata/multipart-mixed.eml")
	require.NoError(b, err)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := imap.NewParsedMessage(literal)
		require.NoError(b, err)
	}
}
