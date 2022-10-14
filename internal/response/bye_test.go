package response

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBye(t *testing.T) {
	raw, filtered := Bye().Strings()
	assert.Equal(t, "* BYE", raw)
	assert.Equal(t, "* BYE", filtered)
}

func TestByeMessage(t *testing.T) {
	raw, filtered := Bye().WithMessage("message").Strings()
	assert.Equal(t, "* BYE message", raw)
	assert.Equal(t, "* BYE message", filtered)
}
