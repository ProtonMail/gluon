package response

import (
	"testing"

	"github.com/ProtonMail/gluon/imap"
	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	assert.Equal(
		t,
		`* LIST (\Noselect) "/" "~/Mail/foo"`,
		List().WithAttributes(imap.NewFlagSet(`\Noselect`)).WithDelimiter("/").WithName(`~/Mail/foo`).String(false),
	)
}

func TestListNilDelimiter(t *testing.T) {
	assert.Equal(
		t,
		`* LIST (\Noselect) NIL "Mail"`,
		List().WithAttributes(imap.NewFlagSet(`\Noselect`)).WithName(`Mail`).String(true),
	)
}
