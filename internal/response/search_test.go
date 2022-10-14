package response

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearch(t *testing.T) {
	raw, filtered := Search(2, 3, 6).Strings()
	assert.Equal(t, `* SEARCH 2 3 6`, raw)
	assert.Equal(t, `* SEARCH 2 3 6`, filtered)
}

func TestSearchEmpty(t *testing.T) {
	raw, filtered := Search().Strings()
	assert.Equal(t, `* SEARCH`, raw)
	assert.Equal(t, `* SEARCH`, filtered)
}
