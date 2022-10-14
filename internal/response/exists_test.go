package response

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExists(t *testing.T) {
	assert.Equal(t, `* 23 EXISTS`, Exists().WithCount(23).String(false))
}
