package rfc822

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanner(t *testing.T) {
	const literal = `this part of the text should be ignored

--longrandomstring

body1

--longrandomstring

body2

--longrandomstring--
`

	scanner, err := NewScanner(strings.NewReader(literal), "longrandomstring")
	require.NoError(t, err)

	parts, err := scanner.ScanAll()
	require.NoError(t, err)

	assert.Equal(t, "\nbody1\n", string(parts[0].Data))
	assert.Equal(t, "\nbody2\n", string(parts[1].Data))

	assert.Equal(t, "\nbody1\n", literal[parts[0].Offset:parts[0].Offset+len(parts[0].Data)])
	assert.Equal(t, "\nbody2\n", literal[parts[1].Offset:parts[1].Offset+len(parts[1].Data)])
}

func TestScannerNested(t *testing.T) {
	const literal = `This is the preamble.  It is to be ignored, though it 
is a handy place for mail composers to include an 
explanatory note to non-MIME compliant readers. 
--simple boundary 
Content-type: multipart/mixed; boundary="nested boundary" 

This is the preamble.  It is to be ignored, though it 
is a handy place for mail composers to include an 
explanatory note to non-MIME compliant readers. 
--nested boundary 
Content-type: text/plain; charset=us-ascii

This part does not end with a linebreak.
--nested boundary 
Content-type: text/plain; charset=us-ascii

This part does end with a linebreak.

--nested boundary--
--simple boundary 
Content-type: text/plain; charset=us-ascii

This part does end with a linebreak.

--simple boundary-- 
This is the epilogue.  It is also to be ignored.
`

	scanner, err := NewScanner(strings.NewReader(literal), "simple boundary")
	require.NoError(t, err)

	parts, err := scanner.ScanAll()
	require.NoError(t, err)

	assert.Equal(t, `Content-type: multipart/mixed; boundary="nested boundary" 

This is the preamble.  It is to be ignored, though it 
is a handy place for mail composers to include an 
explanatory note to non-MIME compliant readers. 
--nested boundary 
Content-type: text/plain; charset=us-ascii

This part does not end with a linebreak.
--nested boundary 
Content-type: text/plain; charset=us-ascii

This part does end with a linebreak.

--nested boundary--`, string(parts[0].Data))
	assert.Equal(t, `Content-type: text/plain; charset=us-ascii

This part does end with a linebreak.
`, string(parts[1].Data))
}

func TestScannerNoFinalLinebreak(t *testing.T) {
	const literal = `--nested boundary 
Content-type: text/plain; charset=us-ascii

This part does not end with a linebreak.
--nested boundary 
Content-type: text/plain; charset=us-ascii

This part does end with a linebreak.

--nested boundary--`

	scanner, err := NewScanner(strings.NewReader(literal), "nested boundary")
	require.NoError(t, err)

	parts, err := scanner.ScanAll()
	require.NoError(t, err)

	assert.Equal(t, `Content-type: text/plain; charset=us-ascii

This part does not end with a linebreak.`, string(parts[0].Data))
	assert.Equal(t, `Content-type: text/plain; charset=us-ascii

This part does end with a linebreak.
`, string(parts[1].Data))
}
