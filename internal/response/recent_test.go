package response

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecent(t *testing.T) {
	assert.Equal(t, `* 5 RECENT`, Recent().WithCount(5).String())
}
