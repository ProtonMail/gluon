package response

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContinuation(t *testing.T) {
	assert.Equal(t, "+ (*_*)", Continuation().String())
}
