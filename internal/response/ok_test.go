package response

import (
	"testing"

	"github.com/ProtonMail/gluon/imap"

	"github.com/stretchr/testify/assert"
)

func TestOkUntagged(t *testing.T) {
	assert.Equal(t, `* OK`, Ok().String())
}

func TestOkTagged(t *testing.T) {
	assert.Equal(t, `tag OK (^_^)`, Ok("tag").String())
}

func TestOkUnseen(t *testing.T) {
	assert.Equal(t, `* OK [UNSEEN 17]`, Ok().WithItems(ItemUnseen(17)).String())
}

func TestOkPermanentFlags(t *testing.T) {
	assert.Equal(t, `* OK [PERMANENTFLAGS (\Deleted)]`, Ok().WithItems(ItemPermanentFlags(imap.NewFlagSet(`\Deleted`))).String())
}

func TestOkUIDNext(t *testing.T) {
	assert.Equal(t, `* OK [UIDNEXT 4392]`, Ok().WithItems(ItemUIDNext(4392)).String())
}

func TestOkUIDValidity(t *testing.T) {
	assert.Equal(t, `* OK [UIDVALIDITY 3857529045]`, Ok().WithItems(ItemUIDValidity(3857529045)).String())
}

func TestOkReadOnly(t *testing.T) {
	assert.Equal(t, `* OK [READ-ONLY]`, Ok().WithItems(ItemReadOnly()).String())
}
