package response

import (
	"testing"

	"github.com/ProtonMail/gluon/imap"
	"github.com/stretchr/testify/assert"
)

func TestOkUntagged(t *testing.T) {
	raw, filtered := Ok().Strings()
	assert.Equal(t, `* OK`, raw)
	assert.Equal(t, `* OK`, filtered)
}

func TestOkTagged(t *testing.T) {
	raw, filtered := Ok("tag").Strings()
	assert.Equal(t, `tag OK`, raw)
	assert.Equal(t, `tag OK`, filtered)
}

func TestOkUnseen(t *testing.T) {
	raw, filtered := Ok().WithItems(ItemUnseen(17)).Strings()
	assert.Equal(t, `* OK [UNSEEN 17]`, raw)
	assert.Equal(t, `* OK [UNSEEN 17]`, filtered)
}

func TestOkPermanentFlags(t *testing.T) {
	raw, filtered := Ok().WithItems(ItemPermanentFlags(imap.NewFlagSet(`\Deleted`))).Strings()
	assert.Equal(t, `* OK [PERMANENTFLAGS (\Deleted)]`, raw)
	assert.Equal(t, `* OK [PERMANENTFLAGS (\Deleted)]`, filtered)
}

func TestOkUIDNext(t *testing.T) {
	raw, filtered := Ok().WithItems(ItemUIDNext(4392)).Strings()
	assert.Equal(t, `* OK [UIDNEXT 4392]`, raw)
	assert.Equal(t, `* OK [UIDNEXT 4392]`, filtered)
}

func TestOkUIDValidity(t *testing.T) {
	raw, filtered := Ok().WithItems(ItemUIDValidity(3857529045)).Strings()
	assert.Equal(t, `* OK [UIDVALIDITY 3857529045]`, raw)
	assert.Equal(t, `* OK [UIDVALIDITY 3857529045]`, filtered)
}

func TestOkReadOnly(t *testing.T) {
	raw, filtered := Ok().WithItems(ItemReadOnly()).Strings()
	assert.Equal(t, `* OK [READ-ONLY]`, raw)
	assert.Equal(t, `* OK [READ-ONLY]`, filtered)
}
