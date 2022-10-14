package response

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoUntagged(t *testing.T) {
	raw, filtered := No().Strings()
	assert.Equal(t, "* NO", raw)
	assert.Equal(t, "* NO", filtered)
}

func TestNoTagged(t *testing.T) {
	raw, filtered := No("tag").Strings()
	assert.Equal(t, "tag NO", raw)
	assert.Equal(t, "tag NO", filtered)
}

func TestNoError(t *testing.T) {
	raw, filtered := No("tag").WithError(errors.New("erroooooor")).Strings()
	assert.Equal(t, "tag NO erroooooor", raw)
	assert.Equal(t, "tag NO erroooooor", filtered)
}

func TestNoTryCreate(t *testing.T) {
	raw, filtered := No("tag").WithItems(ItemTryCreate()).WithError(errors.New("erroooooor")).Strings()
	assert.Equal(t, "tag NO [TRYCREATE] erroooooor", raw)
	assert.Equal(t, "tag NO [TRYCREATE] erroooooor", filtered)
}
