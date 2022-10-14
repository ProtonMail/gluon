package response

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBadUntagged(t *testing.T) {
	raw, filtered := Bad().Strings()
	assert.Equal(t, "* BAD", raw)
	assert.Equal(t, "* BAD", filtered)
}

func TestBadTagged(t *testing.T) {
	raw, filtered := Bad("tag").Strings()
	assert.Equal(t, "tag BAD", raw)
	assert.Equal(t, "tag BAD", filtered)
}

func TestBadError(t *testing.T) {
	raw, filtered := Bad("tag").WithError(errors.New("erroooooor")).Strings()
	assert.Equal(t, "tag BAD erroooooor", raw)
	assert.Equal(t, "tag BAD erroooooor", filtered)
}
