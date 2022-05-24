package response

import (
	"testing"

	"github.com/ProtonMail/gluon/imap"

	"github.com/stretchr/testify/assert"
)

func TestFetch(t *testing.T) {
	assert.Equal(
		t,
		`* 23 FETCH (FLAGS (\Seen) RFC822.SIZE 44827)`,
		Fetch(23).
			WithItems(ItemFlags(imap.NewFlagSet(`\Seen`)), ItemRFC822Size(44827)).
			String(),
	)
}
