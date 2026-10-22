package response

import (
	"math"
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

func TestItemBodyPartialCountOverflow(t *testing.T) {
	assert.Equal(
		t,
		"BODY[]<1> {4}\r\nello",
		ItemBodyLiteral("", []byte("Hello")).WithPartial(1, math.MaxInt).String(),
	)
}
