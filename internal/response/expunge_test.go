package response

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpunge(t *testing.T) {
	assert.Equal(t, `* 23 EXPUNGE`, Expunge(23).String())
}
