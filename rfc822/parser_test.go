package rfc822

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
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

	section := Parse([]byte(literal))

	assert.Equal(t, literal, string(section.Literal()))

	{
		part, err := section.Part(1)
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

--nested boundary--`, string(part.Literal()))
	}

	{
		part, err := section.Part(2)
		require.NoError(t, err)
		assert.Equal(t, `Content-type: text/plain; charset=us-ascii

This part does end with a linebreak.
`, string(part.Literal()))
	}

	{
		part, err := section.Part(1, 1)
		require.NoError(t, err)
		assert.Equal(t, `Content-type: text/plain; charset=us-ascii

This part does not end with a linebreak.`, string(part.Literal()))
	}

	{
		part, err := section.Part(1, 2)
		require.NoError(t, err)
		assert.Equal(t,
			`Content-type: text/plain; charset=us-ascii

This part does end with a linebreak.
`, string(part.Literal()))
	}
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

	section := Parse([]byte(literal))

	assert.Equal(t, literal, string(section.Literal()))

	{
		part, err := section.Part(1)
		require.NoError(t, err)
		assert.Equal(t, `Content-type: text/plain; charset=us-ascii

This part does not end with a linebreak.`, string(part.Literal()))
	}

	{
		part, err := section.Part(2)
		require.NoError(t, err)
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
--embedded-boundary--`, string(part.Literal()))
	}

	{
		part, err := section.Part(2, 1)
		require.NoError(t, err)
		assert.Equal(t, `Content-type: text/plain; charset=us-ascii

This part is embedded

--
From me`, string(part.Literal()))
	}

	{
		part, err := section.Part(2, 2)
		require.NoError(t, err)
		assert.Equal(t, `Content-type: text/plain; charset=us-ascii

This part is also embedded`, string(part.Literal()))
	}
}

func TestParseSpaceLineHeaderMessage(t *testing.T) {
	const literal = `Content-Type: text/plain
Date: Thu, 03 Sep 2020 16:47:43 +0000 (UTC)
Subject: Sometimes
 
 header fields can be long and contain space line :shrug:
From: Dad <dadjokes@arethebest.com>
To: Ships <navy@withbarcode.no>

Why does the Norway navy have bar codes on the side of their ships?

So when they com back to port they can

Scandinavian
`

	section := Parse([]byte(literal))

	assert.Equal(t, literal, string(section.Literal()))

	assert.Equal(t, `Content-Type: text/plain
Date: Thu, 03 Sep 2020 16:47:43 +0000 (UTC)
Subject: Sometimes
 
 header fields can be long and contain space line :shrug:
From: Dad <dadjokes@arethebest.com>
To: Ships <navy@withbarcode.no>

`,
		string(section.Header()),
	)

	assert.Equal(t, `Why does the Norway navy have bar codes on the side of their ships?

So when they com back to port they can

Scandinavian
`,
		string(section.Body()),
	)
}

func TestCarriageReturnHandling(t *testing.T) {
	const literal = "Content-Type: multipart/alternative; boundary=\"------------62DCF50B21CF279F489F0184\"\r\n\r\n\r\n" +
		"--------------62DCF50B21CF279F489F0184\r\nContent-Type: text/plain; charset=utf-8; format=flowed\r\n" +
		"Content-Transfer-Encoding: 7bit\r\n\r\n*this */is**/_html_\r\n**\r\n\r\n--------------62DCF50B21CF279F489F0184\r\n" +
		"Content-Type: text/html; charset=utf-8\r\nContent-Transfer-Encoding: 7bit\r\n<foo></foo>\r\n--------------62DCF50B21CF279F489F0184--\r\n"

	section := Parse([]byte(literal))

	assert.Equal(t, literal, string(section.Literal()))
	{
		part, err := section.Part(1)
		require.NoError(t, err)
		assert.Equal(t, "Content-Type: text/plain; charset=utf-8; format=flowed\r\nContent-Transfer-Encoding: 7bit\r\n\r\n*this */is**/_html_\r\n**\r\n", string(part.Literal()))

	}
}

func TestSectionWalk(t *testing.T) {
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

	section := Parse([]byte(literal))

	var parts [][]byte

	require.NoError(t, section.Walk(func(part *Section) error {
		children, err := part.Children()
		if err != nil {
			return err
		}

		if len(children) == 0 {
			parts = append(parts, part.Body())
		}

		return nil
	}))

	assert.Equal(t, [][]byte{
		[]byte("This part does not end with a linebreak."),
		[]byte("This part does end with a linebreak.\n"),
		[]byte("This part does end with a linebreak.\n"),
	}, parts)
}

func TestSectionDecodedBody(t *testing.T) {
	const literal = `From: Sender <sender@pm.me>
To: Receiver <receiver@pm.me>
Content-Transfer-Encoding: base64

Ym9keQ==
`

	body, err := Parse([]byte(literal)).DecodedBody()
	require.NoError(t, err)

	assert.Equal(t, []byte("body"), body)
}

func FuzzParseDec(f *testing.F) {
	f.Add([]byte(`From: Sender <sender@pm.me>
	To: Receiver <receiver@pm.me>
	Content-Transfer-Encoding: base64
	
	Ym9keQ==
	`))

	f.Add([]byte("Content-Type: multipart/alternative; boundary=\"------------62DCF50B21CF279F489F0184\"\r\n\r\n\r\n" +
		"--------------62DCF50B21CF279F489F0184\r\nContent-Type: text/plain; charset=utf-8; format=flowed\r\n" +
		"Content-Transfer-Encoding: 7bit\r\n\r\n*this */is**/_html_\r\n**\r\n\r\n--------------62DCF50B21CF279F489F0184\r\n" +
		"Content-Type: text/html; charset=utf-8\r\nContent-Transfer-Encoding: 7bit\r\n<foo></foo>\r\n--------------62DCF50B21CF279F489F0184--\r\n"))

	f.Fuzz(func(t *testing.T, inputData []byte) {

		_, _ = Parse(inputData).DecodedBody()
	})
}
