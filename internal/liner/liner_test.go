package liner

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestLinerRead(t *testing.T) {
	uuid.SetRand(new(rander))

	l := New(strings.NewReader(("tag1 login abcde pass\r\n")))
	called := 0

	line, err := l.Read(func() error { called++; return nil })
	assert.NoError(t, err)
	assert.Equal(t, 0, called)
	assert.Equal(t, "tag1 login abcde pass\r\n", string(line))
}

func TestLinerReadOneLiteral(t *testing.T) {
	uuid.SetRand(new(rander))

	l := New(strings.NewReader(("tag1 login {5}\r\nabcde pass\r\n")))
	called := 0

	line, err := l.Read(func() error { called++; return nil })
	assert.NoError(t, err)
	assert.Equal(t, 1, called)
	assert.Equal(t, "tag1 login {5}\r\nabcde pass\r\n", string(line))
}

func TestLinerReadTwoLiterals(t *testing.T) {
	uuid.SetRand(new(rander))

	l := New(strings.NewReader(("tag1 login {5}\r\nabcde {4}\r\npass\r\n")))
	called := 0

	line, err := l.Read(func() error { called++; return nil })
	assert.NoError(t, err)
	assert.Equal(t, 2, called)
	assert.Equal(t, "tag1 login {5}\r\nabcde {4}\r\npass\r\n", string(line))
}

func TestLinerReadMultilineLiteral(t *testing.T) {
	uuid.SetRand(new(rander))

	l := New(strings.NewReader(("tag1 login {15}\r\nabcde\r\nabcdefgh pass\r\n")))
	called := 0

	line, err := l.Read(func() error { called++; return nil })
	assert.NoError(t, err)
	assert.Equal(t, 1, called)
	assert.Equal(t, "tag1 login {15}\r\nabcde\r\nabcdefgh pass\r\n", string(line))
}

type rander byte

func (r *rander) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = byte(*r)
		(*r)++
	}

	return len(p), nil
}
