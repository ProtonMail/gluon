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
		List(false).WithAttributes(imap.NewFlagSet(`\Noselect`)).WithDelimiter("/").WithName(`~/Mail/foo`).String(),
	)

	assert.Equal(
		t,
		`* LSUB (\Noselect) "/" "~/Mail/foo"`,
		List(true).WithAttributes(imap.NewFlagSet(`\Noselect`)).WithDelimiter("/").WithName(`~/Mail/foo`).String(),
	)
}

func TestListNilDelimiter(t *testing.T) {
	assert.Equal(
		t,
		`* LIST (\Noselect) NIL "Mail"`,
		List(false).WithAttributes(imap.NewFlagSet(`\Noselect`)).WithName(`Mail`).String(),
	)

	assert.Equal(
		t,
		`* LSUB (\Noselect) NIL "Mail"`,
		List(true).WithAttributes(imap.NewFlagSet(`\Noselect`)).WithName(`Mail`).String(),
	)
}
