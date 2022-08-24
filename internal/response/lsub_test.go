package response

import (
	"testing"

	"github.com/ProtonMail/gluon/imap"
	"github.com/stretchr/testify/assert"
)

func TestLsub(t *testing.T) {
	assert.Equal(
		t,
		`* LSUB (\Noselect) "/" "~/Mail/foo"`,
		Lsub().WithAttributes(imap.NewFlagSet(`\Noselect`)).WithDelimiter("/").WithName(`~/Mail/foo`).String(),
	)
}

func TestLsubNilDelimiter(t *testing.T) {
	assert.Equal(
		t,
		`* LSUB (\Noselect) NIL "Mail"`,
		Lsub().WithAttributes(imap.NewFlagSet(`\Noselect`)).WithName(`Mail`).String(),
	)
}
