package response

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestItemBodyText(t *testing.T) {
	raw, filtered := ItemBodyLiteral("TEXT", []byte("Hello Joe, do you think we can meet at 3:30 tomorrow?\r\n")).Strings()
	assert.Equal(t, "BODY[TEXT] {55}\r\nHello Joe, do you think we can meet at 3:30 tomorrow?\r\n", raw)
	assert.Equal(t, "", filtered)
}
