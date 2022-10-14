package response

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatus(t *testing.T) {
	raw, filtered := Status().
		WithMailbox(`blurdybloop`).
		WithItems(ItemMessages(231)).
		WithItems(ItemUIDNext(44292)).
		Strings()
	assert.Equal(t, `* STATUS "blurdybloop" (MESSAGES 231 UIDNEXT 44292)`, raw)
	assert.Equal(t, `* STATUS "blurdybloop" (MESSAGES 231 UIDNEXT 44292)`, filtered)
}
