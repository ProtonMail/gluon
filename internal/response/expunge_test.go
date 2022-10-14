package response

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpunge(t *testing.T) {
	raw, filtered := Expunge(23).Strings()
	assert.Equal(t, `* 23 EXPUNGE`, raw)
	assert.Equal(t, `* 23 EXPUNGE`, filtered)
}
