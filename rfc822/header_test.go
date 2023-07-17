package rfc822

import (
	"strings"
	"testing"

	"github.com/bradenaw/juniper/xslices"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const literal = "To: somebody\r\nFrom: somebody else\r\nSubject: this is\r\n\ta multiline field\r\nFrom: duplicate entry\r\n\r\n"

func TestHeader_New(t *testing.T) {
	// Empty headers are empty.
	header, err := NewHeader(nil)
	require.NoError(t, err)
	assert.Equal(t, "", string(header.Raw()))

	// But empty headers can be added to.
	header.Set("To", "someone@pm.me")
	assert.Equal(t, "To: someone@pm.me\r\n", string(header.Raw()))
}

func TestHeader_Raw(t *testing.T) {
	header, err := NewHeader([]byte(literal))
	require.NoError(t, err)
	assert.Equal(t, literal, string(header.Raw()))
}

func TestHeader_Has(t *testing.T) {
	const literal = "To: somebody\r\nFrom: somebody else\r\nSubject: this is\r\n\ta multiline field\r\nFrom: duplicate entry\r\nReferences:\r\n\t <foo@bar.com>\r\n\r\n"

	header, err := NewHeader([]byte(literal))
	require.NoError(t, err)

	assert.Equal(t, true, header.Has("To"))
	assert.Equal(t, true, header.Has("to"))
	assert.Equal(t, false, header.Has("Too"))
	assert.Equal(t, true, header.Has("From"))
	assert.Equal(t, true, header.Has("from"))
	assert.Equal(t, false, header.Has("fromm"))
	assert.Equal(t, true, header.Has("Subject"))
	assert.Equal(t, true, header.Has("subject"))
	assert.Equal(t, false, header.Has("subjectt"))
	assert.Equal(t, true, header.Has("References"))
}

func TestHeader_Get(t *testing.T) {
	const literal = "To: somebody\r\nFrom: somebody else\r\nSubject: this is\r\n\ta multiline field\r\nFrom: duplicate entry\r\nReferences:\r\n\t <foo@bar.com>\r\n\r\n"

	header, err := NewHeader([]byte(literal))
	require.NoError(t, err)

	assert.Equal(t, "somebody", header.Get("To"))
	assert.Equal(t, "somebody", header.Get("to"))
	assert.Equal(t, "somebody else", header.Get("From"))
	assert.Equal(t, "somebody else", header.Get("from"))
	assert.Equal(t, "this is a multiline field", header.Get("Subject"))
	assert.Equal(t, "this is a multiline field", header.Get("subject"))
	assert.Equal(t, "<foo@bar.com>", header.Get("References"))
}

func TestHeader_GetRaw(t *testing.T) {
	header, err := NewHeader([]byte(literal))
	require.NoError(t, err)

	assert.Equal(t, []byte("somebody\r\n"), header.GetRaw("To"))
	assert.Equal(t, []byte("somebody else\r\n"), header.GetRaw("From"))
	assert.Equal(t, []byte("this is\r\n\ta multiline field\r\n"), header.GetRaw("Subject"))
}

func TestHeader_GetLine(t *testing.T) {
	header, err := NewHeader([]byte(literal))
	require.NoError(t, err)

	assert.Equal(t, []byte("To: somebody\r\n"), header.GetLine("To"))
	assert.Equal(t, []byte("From: somebody else\r\n"), header.GetLine("From"))
	assert.Equal(t, []byte("Subject: this is\r\n\ta multiline field\r\n"), header.GetLine("Subject"))
}

func TestHeader_Set(t *testing.T) {
	header, err := NewHeader([]byte(literal))
	require.NoError(t, err)

	assert.Equal(t, "somebody", header.Get("To"))
	header.Set("To", "who is this?")
	assert.Equal(t, "who is this?", header.Get("To"))

	assert.Equal(t, "somebody else", header.Get("From"))
	header.Set("From", "who else is this?")
	assert.Equal(t, "who else is this?", header.Get("From"))
}

func TestHeader_SetNew(t *testing.T) {
	header, err := NewHeader([]byte(literal))
	require.NoError(t, err)

	header.Set("Something", "something new...")
	assert.Equal(t, "something new...", header.Get("Something"))

	assert.Equal(t, "Something: something new...\r\nTo: somebody\r\nFrom: somebody else\r\nSubject: this is\r\n\ta multiline field\r\nFrom: duplicate entry\r\n\r\n", string(header.Raw()))

	header.Set("Else", "another...")
	assert.Equal(t, "another...", header.Get("Else"))

	assert.Equal(t, "Else: another...\r\nSomething: something new...\r\nTo: somebody\r\nFrom: somebody else\r\nSubject: this is\r\n\ta multiline field\r\nFrom: duplicate entry\r\n\r\n", string(header.Raw()))
}

func TestHeader_Del(t *testing.T) {
	header, err := NewHeader([]byte(literal))
	require.NoError(t, err)

	header.Del("From")
	assert.Equal(t, "To: somebody\r\nSubject: this is\r\n\ta multiline field\r\nFrom: duplicate entry\r\n\r\n", string(header.Raw()))

	header.Del("To")
	assert.Equal(t, "Subject: this is\r\n\ta multiline field\r\nFrom: duplicate entry\r\n\r\n", string(header.Raw()))

	header.Del("Subject")
	assert.Equal(t, "From: duplicate entry\r\n\r\n", string(header.Raw()))

	header.Del("From")
	assert.Equal(t, "\r\n", string(header.Raw()))
}

func TestHeader_Fields(t *testing.T) {
	header, err := NewHeader([]byte(literal))
	require.NoError(t, err)

	assert.Equal(t, "To: somebody\r\n\r\n", string(header.Fields([]string{"To"})))
	assert.Equal(t, "From: somebody else\r\nFrom: duplicate entry\r\n\r\n", string(header.Fields([]string{"From"})))
	assert.Equal(t, "To: somebody\r\nFrom: somebody else\r\nFrom: duplicate entry\r\n\r\n", string(header.Fields([]string{"To", "From"})))
	assert.Equal(t, "To: somebody\r\nSubject: this is\r\n\ta multiline field\r\n\r\n", string(header.Fields([]string{"To", "Subject"})))
}

func TestHeader_FieldsNot(t *testing.T) {
	header, err := NewHeader([]byte(literal))
	require.NoError(t, err)

	assert.Equal(t, "To: somebody\r\n\r\n", string(header.FieldsNot([]string{"From", "Subject"})))
	assert.Equal(t, "From: somebody else\r\nFrom: duplicate entry\r\n\r\n", string(header.FieldsNot([]string{"To", "Subject"})))
	assert.Equal(t, "To: somebody\r\nFrom: somebody else\r\nFrom: duplicate entry\r\n\r\n", string(header.FieldsNot([]string{"Subject"})))
	assert.Equal(t, "To: somebody\r\nSubject: this is\r\n\ta multiline field\r\n\r\n", string(header.FieldsNot([]string{"From"})))
}

func TestHeader_Entries(t *testing.T) {
	var lines [][]string

	header, err := NewHeader([]byte(literal))
	require.NoError(t, err)

	header.Entries(func(key, val string) {
		lines = append(lines, []string{key, val})
	})

	assert.Equal(t, [][]string{
		{"To", "somebody"},
		{"From", "somebody else"},
		{"Subject", "this is a multiline field"},
		{"From", "duplicate entry"},
	}, lines)
}

func TestParseHeader(t *testing.T) {
	header, err := NewHeader([]byte(literal))
	require.NoError(t, err)

	assert.Equal(t, [][]byte{
		[]byte("To: somebody\r\n"),
		[]byte("From: somebody else\r\n"),
		[]byte("Subject: this is\r\n\ta multiline field\r\n"),
		[]byte("From: duplicate entry\r\n"),
		[]byte("\r\n"),
	}, header.getLines())
}

func TestParseHeaderFoldedLine(t *testing.T) {
	const literal = "To:\r\n\tsomebody\r\nFrom: \r\n someone\r\n\r\n"

	header, err := NewHeader([]byte(literal))
	require.NoError(t, err)

	assert.Equal(t, [][]byte{
		[]byte("To:\r\n\tsomebody\r\n"),
		[]byte("From: \r\n someone\r\n"),
		[]byte("\r\n"),
	}, header.getLines())
}

func TestParseHeaderMultilineFilename(t *testing.T) {
	const literal = "Content-Type: application/msword; name=\"this is a very long\n filename.doc\""

	header, err := NewHeader([]byte(literal))
	require.NoError(t, err)

	assert.Equal(t, [][]byte{
		[]byte("Content-Type: application/msword; name=\"this is a very long\n filename.doc\""),
	}, header.getLines())
}

func TestParseHeaderMultilineFilenameWithColon(t *testing.T) {
	const literal = "Content-Type: application/msword; name=\"this is a very long\n filename: too long.doc\""

	header, err := NewHeader([]byte(literal))
	require.NoError(t, err)

	assert.Equal(t, [][]byte{
		[]byte("Content-Type: application/msword; name=\"this is a very long\n filename: too long.doc\""),
	}, header.getLines())
}

func TestParseHeaderMultilineFilenameWithColonAndNewline(t *testing.T) {
	const literal = "Content-Type: application/msword; name=\"this is a very long\n filename: too long.doc\"\n"

	header, err := NewHeader([]byte(literal))
	require.NoError(t, err)

	assert.Equal(t, [][]byte{
		[]byte("Content-Type: application/msword; name=\"this is a very long\n filename: too long.doc\"\n"),
	}, header.getLines())
}

func TestParseHeaderMultilineIndent(t *testing.T) {
	const literal = "Subject: a very\r\n\tlong: line with a colon and indent\r\n \r\n and space line\r\nFrom: sender\r\n"

	header, err := NewHeader([]byte(literal))
	require.NoError(t, err)

	assert.Equal(t, [][]byte{
		[]byte("Subject: a very\r\n\tlong: line with a colon and indent\r\n \r\n and space line\r\n"),
		[]byte("From: sender\r\n"),
	}, header.getLines())
}

func TestParseHeaderMultipleMultilineFilenames(t *testing.T) {
	const literal = `Content-Type: application/msword; name="=E5=B8=B6=E6=9C=89=E5=A4=96=E5=9C=8B=E5=AD=97=E7=AC=A6=E7=9A=84=E9=99=84=E4=
 =BB=B6.DOC"
Content-Transfer-Encoding: base64
Content-Disposition: attachment; filename="=E5=B8=B6=E6=9C=89=E5=A4=96=E5=9C=8B=E5=AD=97=E7=AC=A6=E7=9A=84=E9=99=84=E4=
 =BB=B6.DOC"
Content-ID: <>
`

	header, err := NewHeader([]byte(literal))
	require.NoError(t, err)

	assert.Equal(t, [][]byte{
		[]byte("Content-Type: application/msword; name=\"=E5=B8=B6=E6=9C=89=E5=A4=96=E5=9C=8B=E5=AD=97=E7=AC=A6=E7=9A=84=E9=99=84=E4=\n =BB=B6.DOC\"\n"),
		[]byte("Content-Transfer-Encoding: base64\n"),
		[]byte("Content-Disposition: attachment; filename=\"=E5=B8=B6=E6=9C=89=E5=A4=96=E5=9C=8B=E5=AD=97=E7=AC=A6=E7=9A=84=E9=99=84=E4=\n =BB=B6.DOC\"\n"),
		[]byte("Content-ID: <>\n"),
	}, header.getLines())
}

func TestSplitHeaderBody(t *testing.T) {
	const literal = "To: user@pm.me\r\n\r\nhi\r\n"

	header, body := Split([]byte(literal))

	assert.Equal(t, []byte("To: user@pm.me\r\n\r\n"), header)
	assert.Equal(t, []byte("hi\r\n"), body)
}

func TestSplitHeaderBodyNoBody(t *testing.T) {
	const literal = "To: user@pm.me\r\n\r\n"

	header, body := Split([]byte(literal))

	assert.Equal(t, []byte("To: user@pm.me\r\n\r\n"), header)
	assert.Equal(t, []byte(""), body)
}

func TestSplitHeaderBodyOnlyHeader(t *testing.T) {
	const literal = "To: user@pm.me\r\n"

	header, body := Split([]byte(literal))

	assert.Equal(t, []byte("To: user@pm.me\r\n"), header)
	assert.Equal(t, []byte(""), body)
}

func TestSplitHeaderBodyOnlyHeaderNoNewline(t *testing.T) {
	const literal = "To: user@pm.me"

	header, body := Split([]byte(literal))

	assert.Equal(t, []byte("To: user@pm.me"), header)
	assert.Equal(t, []byte(""), body)
}

func TestSetHeaderValue(t *testing.T) {
	const literal = "To: user@pm.me"

	// Create a clone so we can test this with mutable memory.
	literalBytes := xslices.Clone([]byte(literal))

	newHeader, err := SetHeaderValue(literalBytes, "foo", "bar")
	require.NoError(t, err)

	assert.Equal(t, newHeader, []byte("Foo: bar\r\nTo: user@pm.me"))
	// Ensure the original data wasn't modified.
	assert.Equal(t, literalBytes, []byte(literal))
}

func TestHeader_Erase(t *testing.T) {
	literal := []byte("Subject: this is\r\n\ta multiline field\r\nFrom: duplicate entry\r\nReferences:\r\n\t <foo@bar.com>\r\n\r\n")
	literalWithoutSubject := []byte("From: duplicate entry\r\nReferences:\r\n\t <foo@bar.com>\r\n\r\n")
	literalWithoutFrom := []byte("Subject: this is\r\n\ta multiline field\r\nReferences:\r\n\t <foo@bar.com>\r\n\r\n")
	literalWithoutReferences := []byte("Subject: this is\r\n\ta multiline field\r\nFrom: duplicate entry\r\n\r\n")

	{
		newLiteral, err := EraseHeaderValue(literal, "Subject")
		require.NoError(t, err)
		assert.Equal(t, literalWithoutSubject, newLiteral)
	}
	{
		newLiteral, err := EraseHeaderValue(literal, "From")
		require.NoError(t, err)
		assert.Equal(t, literalWithoutFrom, newLiteral)
	}
	{
		newLiteral, err := EraseHeaderValue(literal, "References")
		require.NoError(t, err)
		assert.Equal(t, literalWithoutReferences, newLiteral)
	}
	{
		newLiteral, err := EraseHeaderValue(literal, "ThisKeyDoesNotExist")
		require.NoError(t, err)
		assert.Equal(t, literal, newLiteral)
	}
}

func TestHeader_SubjectWithRandomQuote(t *testing.T) {
	raw := lines(`Subject: All " your " random " brackets " ' ' : belong to us () {}`,
		`Date: Sun, 30 Jan 2000 11:49:30 +0700`,
		`Content-Type: multipart/alternative; boundary="----=_BOUNDARY_"`)

	header, err := NewHeader(raw)
	require.NoError(t, err)

	require.Equal(
		t,
		`All " your " random " brackets " ' ' : belong to us () {}`,
		header.Get("Subject"),
	)
}

func lines(s ...string) []byte {
	return append([]byte(strings.Join(s, "\r\n")), '\r', '\n')
}

func TestHeader_WithTrailingSpaces(t *testing.T) {
	const literal = `From: Nathaniel Borenstein <nsb@bellcore.com> 
To:  Ned Freed <ned@innosoft.com> 
Subject: Sample message 
MIME-Version: 1.0 
Content-type: multipart/mixed; boundary="simple boundary" 
`

	header, err := NewHeader([]byte(literal))
	require.NoError(t, err)

	require.Equal(t, "Nathaniel Borenstein <nsb@bellcore.com>", header.Get("From"))
	require.Equal(t, "Ned Freed <ned@innosoft.com>", header.Get("To"))
	require.Equal(t, "Sample message", header.Get("Subject"))
	require.Equal(t, "1.0", header.Get("MIME-Version"))
	require.Equal(t, `multipart/mixed; boundary="simple boundary"`, header.Get("Content-type"))
}

func TestHeader_MBoxFormatCausesError(t *testing.T) {
	const literal = `X-Mozilla-Keys:
>From 1637354717149124322@xxx Tue Jun 25 22:52:20 +0000 2019
X-GM-THIRD: 12345
`

	_, err := NewHeader([]byte(literal))
	require.Error(t, err)
}

func TestHeader_EmptyField(t *testing.T) {
	const literal = "X-Mozilla-Keys:\r\nX-GM-THIRD: 12345\r\n"

	header, err := NewHeader([]byte(literal))
	require.NoError(t, err)
	require.Empty(t, header.Get("X-Mozilla-Key"))
	require.Equal(t, "12345", header.Get("X-GM-THIRD"))
}

func TestHeader_MissingLFAfterCRIsError(t *testing.T) {
	const literal = "X-Mozilla-Keys:\rX-GM-THIRD: 12345\r\n"

	_, err := NewHeader([]byte(literal))
	require.Error(t, err)
}

func TestHeader_SingleEmptyField(t *testing.T) {
	header, err := NewHeader([]byte("Content-tYpe:\r")) //Panic
	require.NoError(t, err)

	require.Empty(t, header.Get("Content-Type"))
}

func TestHeader_NoSpaceAfterColonIsValid(t *testing.T) {
	header, err := NewHeader([]byte("Content-tYpe:Foobar\r\n")) //Panic
	require.NoError(t, err)

	require.Equal(t, "Foobar", header.Get("Content-Type"))
}

func TestHeader_WithTabs(t *testing.T) {
	const literal = "From: Bar <bar@bar.com>\n" +
		"Date: 01 Jan 1980 00:00:00 +0000\n" +
		"Subject: Weird header field\n" +
		"To:\t<receiver@pm.test>,\n" +
		" <another@pm.test>\n"

	header, err := NewHeader([]byte(literal))
	require.NoError(t, err)
	require.Equal(t, "<receiver@pm.test>, <another@pm.test>", string(header.Get("To")))
}
