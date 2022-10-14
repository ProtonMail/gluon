package response

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExists(t *testing.T) {
	raw, filtered := Exists().WithCount(23).Strings()
	assert.Equal(t, `* 23 EXISTS`, raw)
	assert.Equal(t, `* 23 EXISTS`, filtered)
}
