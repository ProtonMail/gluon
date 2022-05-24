package response

import (
	"testing"

	"github.com/ProtonMail/gluon/imap"

	"github.com/stretchr/testify/assert"
)

func TestLsub(t *testing.T) {
	assert.Equal(
		t,
		`* LSUB (\NoSelect) "/" "~/Mail/foo"`,
		Lsub().WithAttributes(imap.NewFlagSet(`\NoSelect`)).WithDelimiter("/").WithName(`~/Mail/foo`).String(),
	)
}

func TestLsubNilDelimiter(t *testing.T) {
	assert.Equal(
		t,
		`* LSUB (\NoSelect) NIL "Mail"`,
		Lsub().WithAttributes(imap.NewFlagSet(`\NoSelect`)).WithName(`Mail`).String(),
	)
}
