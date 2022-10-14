package response

import (
	"testing"

	"github.com/ProtonMail/gluon/imap"
	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	raw, filtered := List().WithAttributes(imap.NewFlagSet(`\Noselect`)).WithDelimiter("/").WithName(`~/Mail/foo`).Strings()
	assert.Equal(t, `* LIST (\Noselect) "/" "~/Mail/foo"`, raw)
	assert.Equal(t, `* LIST (\Noselect) "/" "~/Mail/foo"`, filtered)
}

func TestListNilDelimiter(t *testing.T) {
	raw, filtered := List().WithAttributes(imap.NewFlagSet(`\Noselect`)).WithName(`Mail`).Strings()
	assert.Equal(t, `* LIST (\Noselect) NIL "Mail"`, raw)
	assert.Equal(t, `* LIST (\Noselect) NIL "Mail"`, filtered)
}
