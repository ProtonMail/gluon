package response

import (
	"testing"

	"github.com/ProtonMail/gluon/imap"
	"github.com/stretchr/testify/assert"
)

func TestLsub(t *testing.T) {
	raw, filtered := Lsub().WithAttributes(imap.NewFlagSet(`\Noselect`)).WithDelimiter("/").WithName(`~/Mail/foo`).Strings()
	assert.Equal(t, `* LSUB (\Noselect) "/" "~/Mail/foo"`, raw)
	assert.Equal(t, `* LSUB (\Noselect) "/" "~/Mail/foo"`, filtered)
}

func TestLsubNilDelimiter(t *testing.T) {
	raw, filtered := Lsub().WithAttributes(imap.NewFlagSet(`\Noselect`)).WithName(`Mail`).Strings()
	assert.Equal(t, `* LSUB (\Noselect) NIL "Mail"`, raw)
	assert.Equal(t, `* LSUB (\Noselect) NIL "Mail"`, filtered)
}
