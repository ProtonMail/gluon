package response

import (
	"testing"

	"github.com/ProtonMail/gluon/imap"

	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	assert.Equal(
		t,
		`* LIST (\NoSelect) "/" "~/Mail/foo"`,
		List().WithAttributes(imap.NewFlagSet(`\NoSelect`)).WithDelimiter("/").WithName(`~/Mail/foo`).String(),
	)
}

func TestListNilDelimiter(t *testing.T) {
	assert.Equal(
		t,
		`* LIST (\NoSelect) NIL "Mail"`,
		List().WithAttributes(imap.NewFlagSet(`\NoSelect`)).WithName(`Mail`).String(),
	)
}
