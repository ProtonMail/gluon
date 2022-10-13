package rfc822

import (
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

	{
		scanner, err := NewByteScanner([]byte(literal), []byte("longrandomstring"))
		require.NoError(t, err)

		parts := scanner.ScanAll()
		require.Equal(t, len(parts), 2)

		assert.Equal(t, "\nbody1\n", string(parts[0].Data))
		assert.Equal(t, "\nbody2\n", string(parts[1].Data))

		assert.Equal(t, "\nbody1\n", literal[parts[0].Offset:parts[0].Offset+len(parts[0].Data)])
		assert.Equal(t, "\nbody2\n", literal[parts[1].Offset:parts[1].Offset+len(parts[1].Data)])
	}
}

func TestScannerMalformedLineEnding(t *testing.T) {
	// contains an extra \r before new line after boundary
	const literal = "this part of the text should be ignored\r\n--longrandomstring\r\r\n\r\nbody1\r\n\r\n--longrandomstring\r\r\n\r\nbody2\r\n\r\n--longrandomstring--\r\r\n"

	{
		scanner, err := NewByteScanner([]byte(literal), []byte("longrandomstring"))
		require.NoError(t, err)

		parts := scanner.ScanAll()
		require.Equal(t, len(parts), 2)

		assert.Equal(t, "\r\nbody1\r\n", string(parts[0].Data))
		assert.Equal(t, "\r\nbody2\r\n", string(parts[1].Data))

		assert.Equal(t, "\r\nbody1\r\n", literal[parts[0].Offset:parts[0].Offset+len(parts[0].Data)])
		assert.Equal(t, "\r\nbody2\r\n", literal[parts[1].Offset:parts[1].Offset+len(parts[1].Data)])
	}
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

	scanner, err := NewByteScanner([]byte(literal), []byte("simple boundary"))
	require.NoError(t, err)

	parts := scanner.ScanAll()
	require.Equal(t, 2, len(parts))

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
	const literal = `
--nested boundary
Content-type: text/plain; charset=us-ascii

This part does not end with a linebreak.
--nested boundary
Content-type: text/plain; charset=us-ascii

This part does end with a linebreak.

--nested boundary--`

	scanner, err := NewByteScanner([]byte(literal), []byte("nested boundary"))
	require.NoError(t, err)

	parts := scanner.ScanAll()
	require.Equal(t, 2, len(parts))

	assert.Equal(t, `Content-type: text/plain; charset=us-ascii

This part does not end with a linebreak.`, string(parts[0].Data))
	assert.Equal(t, `Content-type: text/plain; charset=us-ascii

This part does end with a linebreak.
`, string(parts[1].Data))
}
