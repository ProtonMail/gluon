package response

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecent(t *testing.T) {
	raw, filtered := Recent().WithCount(5).Strings()
	assert.Equal(t, `* 5 RECENT`, raw)
	assert.Equal(t, `* 5 RECENT`, filtered)
}
