package imap

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmbeddedRFC822WithoutHeader(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "rfc822.eml"))
	require.NoError(t, err)

	parsed, err := NewParsedMessage(b)
	require.NoError(t, err)
	require.NotNil(t, parsed)
}

func TestHeaderOutOfBounds(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "bounds.eml"))
	require.NoError(t, err)

	parsed, err := NewParsedMessage(b)
	require.NoError(t, err)
	require.NotNil(t, parsed)
}

func TestStructureWithRFC822Embedded(t *testing.T) {
	const message = `Content-Type: multipart/mixed;
 boundary=dcd8fbdd2e8a8f95ac2024a5a57b37e2c24da4f0a0006ae059da17cb0e5b
Return-Path: <random-mail2@pm.me>
X-Original-To: random-mail@pm.me
Delivered-To: random-mail@pm.me
References: <>
Subject: Fwd: ISO-8859-1
To: random-mail@pm.me
From: BQA <random-mail2@pm.me>
X-Forwarded-Message-Id: <>
Message-Id: <0f57877e-0003-b600-9e62-8ad2736ec325@gmail.com>
Date: Wed, 2 Jun 2021 14:18:56 +0200
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:78.0) Gecko/20100101
 Thunderbird/78.10.2
Mime-Version: 1.0
In-Reply-To: <>
Content-Language: en-US

--dcd8fbdd2e8a8f95ac2024a5a57b37e2c24da4f0a0006ae059da17cb0e5b
Content-Transfer-Encoding: quoted-printable
Content-Type: text/plain; charset=utf-8

what


--dcd8fbdd2e8a8f95ac2024a5a57b37e2c24da4f0a0006ae059da17cb0e5b
Content-Disposition: attachment; filename=ISO-8859-1.eml
Content-Type: message/rfc822; name=ISO-8859-1.eml

From: random-mail@pm.me
To: random-mail2@pm.me
Content-Type: text/plain; charset=iso-8859-1
Subject: ISO-8859-1

hey there bro

--dcd8fbdd2e8a8f95ac2024a5a57b37e2c24da4f0a0006ae059da17cb0e5b--
`

	parsed, err := NewParsedMessage([]byte(message))
	require.NoError(t, err)
	require.NotNil(t, parsed)

	expected := "((\"text\" \"plain\" (\"charset\" \"utf-8\") NIL NIL \"quoted-printable\" 6 2)(\"message\" \"rfc822\" (\"name\" \"ISO-8859-1.eml\") NIL NIL NIL 127 (NIL \"ISO-8859-1\" ((NIL NIL \"random-mail\" \"pm.me\")) ((NIL NIL \"random-mail\" \"pm.me\")) ((NIL NIL \"random-mail\" \"pm.me\")) ((NIL NIL \"random-mail2\" \"pm.me\")) NIL NIL NIL NIL)(\"text\" \"plain\" (\"charset\" \"iso-8859-1\") NIL NIL NIL 14 1) 6) \"mixed\")"
	require.Equal(t, expected, parsed.Body)
}

func TestParseInvalidCharsInContenType(t *testing.T) {
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
Content-Type: GIF �ɮ�;
Content-Transfer-Encoding: base64
Content-Location: file:///C:/ABWhiz41/images/itinlogo1.gif

Ym9keQ==

--embedded-boundary
Content-type: text/plain; charset=us-ascii

This part is also embedded
--embedded-boundary--
--simple boundary--
This is the epilogue.  It is also to be ignored.
`

	parsed, err := NewParsedMessage([]byte(literal))
	require.NoError(t, err)
	require.NotNil(t, parsed)
}

func TestParseInvalidMimeType(t *testing.T) {
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
Content-Type: application/;
Content-Transfer-Encoding: base64
Content-Location: file:///C:/ABWhiz41/images/itinlogo1.gif

Ym9keQ==

--embedded-boundary
Content-type: text/plain; charset=us-ascii

This part is also embedded
--embedded-boundary--
--simple boundary--
This is the epilogue.  It is also to be ignored.
`

	parsed, err := NewParsedMessage([]byte(literal))
	require.NoError(t, err)
	require.NotNil(t, parsed)
}

func TestParseMessage_GODT_2513(t *testing.T) {
	// This sequence could cause a crash.
	parsed, err := NewParsedMessage([]byte("Content-tYpe: multipArt/0;BoundArY=\"simple boundary\"\n\n--simple boundary\r"))
	require.NoError(t, err)
	require.NotNil(t, parsed)
}

func Fuzz_NewParsedMessage(f *testing.F) {
	in_seed1, _ := os.ReadFile(filepath.Join("testdata", "rfc822.eml"))
	in_seed2, _ := os.ReadFile(filepath.Join("testdata", "bounds.eml"))

	f.Add(in_seed1)
	f.Add(in_seed2)

	f.Fuzz(func(t *testing.T, inputData []byte) {

		_, _ = NewParsedMessage(inputData)
	})
}
