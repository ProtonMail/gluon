package response

import (
	"testing"

	"github.com/ProtonMail/gluon/imap"

	"github.com/stretchr/testify/assert"
)

func TestFlags(t *testing.T) {
	assert.Equal(
		t,
		`* FLAGS (\Answered \Deleted \Draft \Flagged \Seen)`,
		Flags().WithFlags(imap.NewFlagSet(`\Answered`, `\Flagged`, `\Deleted`, `\Seen`, `\Draft`)).String(),
	)
}
