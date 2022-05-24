package response

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBye(t *testing.T) {
	assert.Equal(t, "* BYE (^_^)/~", Bye().String())
}

func TestByeMessage(t *testing.T) {
	assert.Equal(t, "* BYE (^_^)/~ message", Bye().WithMessage("message").String())
}
