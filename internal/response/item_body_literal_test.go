package response

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestItemBodyText(t *testing.T) {
	assert.Equal(
		t,
		"BODY[TEXT] {55}\r\nHello Joe, do you think we can meet at 3:30 tomorrow?\r\n",
		ItemBodyLiteral("TEXT", []byte("Hello Joe, do you think we can meet at 3:30 tomorrow?\r\n")).String(),
	)
}
