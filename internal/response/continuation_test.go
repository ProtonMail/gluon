package response

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContinuation(t *testing.T) {
	raw, filtered := Continuation().Strings()
	assert.Equal(t, "+ Ready", raw)
	assert.Equal(t, "+ Ready", filtered)
}
