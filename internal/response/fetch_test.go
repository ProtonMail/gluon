package response

import (
	"testing"

	"github.com/ProtonMail/gluon/imap"
	"github.com/stretchr/testify/assert"
)

func TestFetch(t *testing.T) {
	raw, filtered := Fetch(23).WithItems(ItemFlags(imap.NewFlagSet(`\Seen`)), ItemRFC822Size(44827)).Strings()
	assert.Equal(t, `* 23 FETCH (FLAGS (\Seen) RFC822.SIZE 44827)`, raw)
	assert.Equal(t, `* 23 FETCH (FLAGS (\Seen) RFC822.SIZE 44827)`, filtered)

	raw, filtered = Fetch(23).WithItems(ItemFlags(imap.NewFlagSet(`\Seen`)), ItemRFC822Text([]byte("foo bar"))).Strings()
	assert.Equal(t, "* 23 FETCH (FLAGS (\\Seen) RFC822.TEXT {7}\r\nfoo bar)", raw)
	assert.Equal(t, "* 23 FETCH (FLAGS (\\Seen) RFC822.TEXT {7})", filtered)
}
