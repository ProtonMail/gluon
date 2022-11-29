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

	line, literals, err := l.Read(func() error { called++; return nil })
	assert.NoError(t, err)
	assert.Equal(t, 0, called)
	assert.Equal(t, "tag1 login abcde pass\r\n", string(line))
	assert.Len(t, literals, 0)
}

func TestLinerReadOneLiteral(t *testing.T) {
	uuid.SetRand(new(rander))

	l := New(strings.NewReader(("tag1 login {5}\r\nabcde pass\r\n")))
	called := 0

	line, literals, err := l.Read(func() error { called++; return nil })
	assert.NoError(t, err)
	assert.Equal(t, 1, called)
	assert.Equal(t, "tag1 login {5}\r\n00010203-0405-4607-8809-0a0b0c0d0e0f pass\r\n", string(line))
	assert.Equal(t, "abcde", string(literals["00010203-0405-4607-8809-0a0b0c0d0e0f"]))
	assert.Len(t, literals, 1)
}

func TestLinerReadTwoLiterals(t *testing.T) {
	uuid.SetRand(new(rander))

	l := New(strings.NewReader(("tag1 login {5}\r\nabcde {4}\r\npass\r\n")))
	called := 0

	line, literals, err := l.Read(func() error { called++; return nil })
	assert.NoError(t, err)
	assert.Equal(t, 2, called)
	assert.Equal(t, "tag1 login {5}\r\n00010203-0405-4607-8809-0a0b0c0d0e0f {4}\r\n10111213-1415-4617-9819-1a1b1c1d1e1f\r\n", string(line))
	assert.Equal(t, "abcde", string(literals["00010203-0405-4607-8809-0a0b0c0d0e0f"]))
	assert.Equal(t, "pass", string(literals["10111213-1415-4617-9819-1a1b1c1d1e1f"]))
	assert.Len(t, literals, 2)
}

func TestLinerReadMultilineLiteral(t *testing.T) {
	uuid.SetRand(new(rander))

	l := New(strings.NewReader(("tag1 login {15}\r\nabcde\r\nabcdefgh pass\r\n")))
	called := 0

	line, literals, err := l.Read(func() error { called++; return nil })
	assert.NoError(t, err)
	assert.Equal(t, 1, called)
	assert.Equal(t, "tag1 login {15}\r\n00010203-0405-4607-8809-0a0b0c0d0e0f pass\r\n", string(line))
	assert.Equal(t, "abcde\r\nabcdefgh", string(literals["00010203-0405-4607-8809-0a0b0c0d0e0f"]))
	assert.Len(t, literals, 1)
}

type rander byte

func (r *rander) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = byte(*r)
		(*r)++
	}

	return len(p), nil
}
