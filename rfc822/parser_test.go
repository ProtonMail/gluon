package rfc822

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseNestedMultipartMixed(t *testing.T) {
	const literal = `From: Nathaniel Borenstein <nsb@bellcore.com> 
To:  Ned Freed <ned@innosoft.com> 
Subject: Sample message 
MIME-Version: 1.0 
Content-type: multipart/mixed; boundary="simple boundary" 

This is the preamble.  It is to be ignored, though it 
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

	section, err := Parse([]byte(literal))
	require.NoError(t, err)

	assert.Equal(t, literal, string(section.Literal()))

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

--nested boundary--`, string(section.Part(1).Literal()))

	assert.Equal(t, `Content-type: text/plain; charset=us-ascii

This part does end with a linebreak.
`, string(section.Part(2).Literal()))

	assert.Equal(t, `Content-type: text/plain; charset=us-ascii

This part does not end with a linebreak.`, string(section.Part(1, 1).Literal()))

	assert.Equal(t,
		`Content-type: text/plain; charset=us-ascii

This part does end with a linebreak.
`, string(section.Part(1, 2).Literal()))
}

func TestParseEmbeddedMessage(t *testing.T) {
	const literal = `From: Nathaniel Borenstein <nsb@bellcore.com> 
To:  Ned Freed <ned@innosoft.com> 
Subject: Sample message 
MIME-Version: 1.0 
Content-type: multipart/mixed; boundary="simple boundary" 

This is the preamble.  It is to be ignored, though it 
is a handy place for mail composers to include an 
explanatory note to non-MIME compliant readers. 
--simple boundary 
Content-type: text/plain; charset=us-ascii

This part does not end with a linebreak.
--simple boundary 
Content-Disposition: attachment; filename=test.eml
Content-Type: message/rfc822; name=test.eml
X-Pm-Content-Encryption: on-import

To: someone
Subject: Fwd: embedded
Content-type: multipart/mixed; boundary="embedded-boundary" 

--embedded-boundary
Content-type: text/plain; charset=us-ascii

This part is embedded

--
From me
--embedded-boundary
Content-type: text/plain; charset=us-ascii

This part is also embedded
--embedded-boundary--
--simple boundary-- 
This is the epilogue.  It is also to be ignored.
`

	section, err := Parse([]byte(literal))
	require.NoError(t, err)

	assert.Equal(t, literal, string(section.Literal()))

	assert.Equal(t, `Content-type: text/plain; charset=us-ascii

This part does not end with a linebreak.`, string(section.Part(1).Literal()))

	assert.Equal(t, `Content-Disposition: attachment; filename=test.eml
Content-Type: message/rfc822; name=test.eml
X-Pm-Content-Encryption: on-import

To: someone
Subject: Fwd: embedded
Content-type: multipart/mixed; boundary="embedded-boundary" 

--embedded-boundary
Content-type: text/plain; charset=us-ascii

This part is embedded

--
From me
--embedded-boundary
Content-type: text/plain; charset=us-ascii

This part is also embedded
--embedded-boundary--`, string(section.Part(2).Literal()))

	assert.Equal(t, `Content-type: text/plain; charset=us-ascii

This part is embedded

--
From me`, string(section.Part(2, 1).Literal()))

	assert.Equal(t, `Content-type: text/plain; charset=us-ascii

This part is also embedded`, string(section.Part(2, 2).Literal()))
}
