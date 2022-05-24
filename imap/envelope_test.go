package imap_test

import (
	"os"
	"testing"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/rfc822"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvelope(t *testing.T) {
	b, err := os.ReadFile("testdata/envelope.eml")
	require.NoError(t, err)

	root, err := rfc822.Parse(b)
	require.NoError(t, err)

	envelope, err := imap.Envelope(root.ParseHeader())
	require.NoError(t, err)

	assert.Equal(t, "(\"Sat, 03 Apr 2021 15:13:53 +0000\" \"this is currently a draft\" ((NIL NIL \"somebody\" \"pm.me\")) ((NIL NIL \"somebody\" \"pm.me\")) ((NIL NIL \"somebody\" \"pm.me\")) ((\"Somebody\" NIL \"somebody\" \"pm.me\")) NIL NIL NIL \"<X9xiWTZnfxfC0wGLBI9t-WEJCOSO_pT67TjlDDKZxzs7TFRCvzCF8lCtqrflZ9n2Z8Ve3rhwYE-vzUGkgOJWaZK4VWMk_WbertE5uklqS8A=@pm.me>\")", envelope)
}
