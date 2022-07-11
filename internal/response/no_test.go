package response

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoUntagged(t *testing.T) {
	assert.Equal(t, "* NO", No().String())
}

func TestNoTagged(t *testing.T) {
	assert.Equal(t, "tag NO", No("tag").String())
}

func TestNoError(t *testing.T) {
	assert.Equal(t, "tag NO erroooooor", No("tag").WithError(errors.New("erroooooor")).String())
}

func TestNoTryCreate(t *testing.T) {
	assert.Equal(t, "tag NO [TRYCREATE] erroooooor", No("tag").WithItems(ItemTryCreate()).WithError(errors.New("erroooooor")).String())
}
