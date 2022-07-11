package response

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBadUntagged(t *testing.T) {
	assert.Equal(t, "* BAD", Bad().String())
}

func TestBadTagged(t *testing.T) {
	assert.Equal(t, "tag BAD", Bad("tag").String())
}

func TestBadError(t *testing.T) {
	assert.Equal(t, "tag BAD erroooooor", Bad("tag").WithError(errors.New("erroooooor")).String())
}
