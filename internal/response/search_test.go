package response

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearch(t *testing.T) {
	assert.Equal(
		t,
		`* SEARCH 2 3 6`,
		Search(2, 3, 6).String(),
	)
}

func TestSearchEmpty(t *testing.T) {
	assert.Equal(
		t,
		`* SEARCH`,
		Search().String(),
	)
}
