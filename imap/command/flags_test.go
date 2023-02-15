package command

import (
	"bytes"
	rfcparser "github.com/ProtonMail/gluon/rfcparser"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParser_ParseFlagList(t *testing.T) {
	values := map[string][]string{
		`(\Answered)`:                {`\Answered`},
		`(\Answered Foo \Something)`: {`\Answered`, `Foo`, `\Something`},
	}

	for input, expected := range values {
		p := rfcparser.NewParser(rfcparser.NewScanner(bytes.NewReader([]byte(input))))
		require.NoError(t, p.Advance())
		v, err := ParseFlagList(p)
		require.NoError(t, err)
		require.Equal(t, expected, v)
	}
}

func TestParser_ParseFlagListInvalid(t *testing.T) {
	inputs := [][]byte{
		[]byte(`()`),
		[]byte(`(\Foo\Bar)`),
		[]byte(`"(\Recent)`),
		[]byte(`(\Foo )`),
		[]byte(`(\Foo`),
	}
	for _, i := range inputs {
		p := rfcparser.NewParser(rfcparser.NewScanner(bytes.NewReader([]byte(i))))
		require.NoError(t, p.Advance())

		_, err := ParseFlagList(p)
		require.Error(t, err)
	}
}
