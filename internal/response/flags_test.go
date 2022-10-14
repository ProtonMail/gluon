package response

import (
	"testing"

	"github.com/ProtonMail/gluon/imap"

	"github.com/stretchr/testify/assert"
)

func TestFlags(t *testing.T) {
	raw, filtered := Flags().WithFlags(imap.NewFlagSet(`\Answered`, `\Flagged`, `\Deleted`, `\Seen`, `\Draft`)).Strings()
	assert.Equal(t, `* FLAGS (\Answered \Deleted \Draft \Flagged \Seen)`, raw)
	assert.Equal(t, `* FLAGS (\Answered \Deleted \Draft \Flagged \Seen)`, filtered)
}
