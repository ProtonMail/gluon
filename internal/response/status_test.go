package response

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatus(t *testing.T) {
	assert.Equal(
		t,
		`* STATUS "blurdybloop" (MESSAGES 231 UIDNEXT 44292)`,
		Status().
			WithMailbox(`blurdybloop`).
			WithItems(ItemMessages(231)).
			WithItems(ItemUIDNext(44292)).
			String(false),
	)
}
