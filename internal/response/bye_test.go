package response

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBye(t *testing.T) {
	assert.Equal(t, "* BYE", Bye().String(false))
}

func TestByeMessage(t *testing.T) {
	assert.Equal(t, "* BYE message", Bye().WithMessage("message").String(false))
}
