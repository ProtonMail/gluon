package rfc5322

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseAddrSpec(t *testing.T) {
	inputs := map[string]string{
		`pete(his account)@silly.test(his host)`: `pete@silly.test`,
		`jdoe@machine.example`:                   `jdoe@machine.example`,
		`john.q.public@example.com`:              `john.q.public@example.com`,
		`user@example.com`:                       `user@example.com`,
		`user@[10.0.0.1]`:                        `user@[10.0.0.1]`,
		`hořejšek@mail.com `:                     `hořejšek@mail.com`,
	}

	for i, e := range inputs {
		t.Run(i, func(t *testing.T) {
			p := newTestRFCParser(i)
			v, err := parseAddrSpec(p)
			require.NoError(t, err)
			require.Equal(t, e, v)
		})
	}
}

func TestParseAngleAddr(t *testing.T) {
	inputs := map[string]string{
		`<pete(his account)@silly.test(his host)>`: `pete@silly.test`,
		`<jdoe@machine.example>`:                   `jdoe@machine.example`,
		`<john.q.public@example.com>`:              `john.q.public@example.com`,
		`<user@example.com>`:                       `user@example.com`,
		`<user@[10.0.0.1]>`:                        `user@[10.0.0.1]`,
		`<hořejšek@mail.com>`:                      `hořejšek@mail.com`,
		`<@foo.com:foo@bar.com>`:                   `foo@bar.com`,
		`<,@foo.com:foo@bar.com>`:                  `foo@bar.com`,
		`<  @foo.com:foo@bar.com>`:                 `foo@bar.com`,
		`<@foo.com,@bar.bar:foo@bar.com>`:          `foo@bar.com`,
		"<@foo.com,\r\n @bar.bar:foo@bar.com>":     `foo@bar.com`,
	}

	for i, e := range inputs {
		t.Run(i, func(t *testing.T) {
			p := newTestRFCParser(i)
			v, err := parseAngleAddr(p)
			require.NoError(t, err)
			require.Equal(t, e, v)
		})
	}
}
