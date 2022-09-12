package rfc822

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const literal = "To: somebody\r\nFrom: somebody else\r\nSubject: this is\r\n\ta multiline field\r\nFrom: duplicate entry\r\n\r\n"

func TestHeader_Raw(t *testing.T) {
	header, err := ParseHeader([]byte(literal))
	require.NoError(t, err)
	assert.Equal(t, literal, string(header.Raw()))
}

func TestHeader_Has(t *testing.T) {
	const literal = "To: somebody\r\nFrom: somebody else\r\nSubject: this is\r\n\ta multiline field\r\nFrom: duplicate entry\r\nReferences:\r\n\t <foo@bar.com>\r\n\r\n"

	header, err := ParseHeader([]byte(literal))
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

	header, err := ParseHeader([]byte(literal))
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
	header, err := ParseHeader([]byte(literal))
	require.NoError(t, err)

	assert.Equal(t, []byte("somebody\r\n"), header.GetRaw("To"))
	assert.Equal(t, []byte("somebody else\r\n"), header.GetRaw("From"))
	assert.Equal(t, []byte("this is\r\n\ta multiline field\r\n"), header.GetRaw("Subject"))
}

func TestHeader_GetLine(t *testing.T) {
	header, err := ParseHeader([]byte(literal))
	require.NoError(t, err)

	assert.Equal(t, []byte("To: somebody\r\n"), header.GetLine("To"))
	assert.Equal(t, []byte("From: somebody else\r\n"), header.GetLine("From"))
	assert.Equal(t, []byte("Subject: this is\r\n\ta multiline field\r\n"), header.GetLine("Subject"))
}

func TestHeader_Set(t *testing.T) {
	header, err := ParseHeader([]byte(literal))
	require.NoError(t, err)

	assert.Equal(t, "somebody", header.Get("To"))
	header.Set("To", "who is this?")
	assert.Equal(t, "who is this?", header.Get("To"))

	assert.Equal(t, "somebody else", header.Get("From"))
	header.Set("From", "who else is this?")
	assert.Equal(t, "who else is this?", header.Get("From"))
}

func TestHeader_SetNew(t *testing.T) {
	header, err := ParseHeader([]byte(literal))
	require.NoError(t, err)

	header.Set("Something", "something new...")
	assert.Equal(t, "something new...", header.Get("Something"))

	assert.Equal(t, "Something: something new...\r\nTo: somebody\r\nFrom: somebody else\r\nSubject: this is\r\n\ta multiline field\r\nFrom: duplicate entry\r\n\r\n", string(header.Raw()))

	header.Set("Else", "another...")
	assert.Equal(t, "another...", header.Get("Else"))

	assert.Equal(t, "Else: another...\r\nSomething: something new...\r\nTo: somebody\r\nFrom: somebody else\r\nSubject: this is\r\n\ta multiline field\r\nFrom: duplicate entry\r\n\r\n", string(header.Raw()))
}

func TestHeader_Del(t *testing.T) {
	header, err := ParseHeader([]byte(literal))
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
	header, err := ParseHeader([]byte(literal))
	require.NoError(t, err)

	assert.Equal(t, "To: somebody\r\n\r\n", string(header.Fields([]string{"To"})))
	assert.Equal(t, "From: somebody else\r\nFrom: duplicate entry\r\n\r\n", string(header.Fields([]string{"From"})))
	assert.Equal(t, "To: somebody\r\nFrom: somebody else\r\nFrom: duplicate entry\r\n\r\n", string(header.Fields([]string{"To", "From"})))
	assert.Equal(t, "To: somebody\r\nSubject: this is\r\n\ta multiline field\r\n\r\n", string(header.Fields([]string{"To", "Subject"})))
}

func TestHeader_FieldsNot(t *testing.T) {
	header, err := ParseHeader([]byte(literal))
	require.NoError(t, err)

	assert.Equal(t, "To: somebody\r\n\r\n", string(header.FieldsNot([]string{"From", "Subject"})))
	assert.Equal(t, "From: somebody else\r\nFrom: duplicate entry\r\n\r\n", string(header.FieldsNot([]string{"To", "Subject"})))
	assert.Equal(t, "To: somebody\r\nFrom: somebody else\r\nFrom: duplicate entry\r\n\r\n", string(header.FieldsNot([]string{"Subject"})))
	assert.Equal(t, "To: somebody\r\nSubject: this is\r\n\ta multiline field\r\n\r\n", string(header.FieldsNot([]string{"From"})))
}

func TestHeader_Entries(t *testing.T) {
	var lines [][]string

	header, err := ParseHeader([]byte(literal))
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
	header, err := ParseHeader([]byte(literal))
	require.NoError(t, err)

	assert.Equal(t, [][]byte{
		[]byte("To: somebody\r\n"),
		[]byte("From: somebody else\r\n"),
		[]byte("Subject: this is\r\n\ta multiline field\r\n"),
		[]byte("From: duplicate entry\r\n"),
		[]byte("\r\n"),
	}, header.lines)
}

func TestParseHeaderFoldedLine(t *testing.T) {
	const literal = "To:\r\n\tsomebody\r\nFrom: \r\n someone\r\n\r\n"

	header, err := ParseHeader([]byte(literal))
	require.NoError(t, err)

	assert.Equal(t, [][]byte{
		[]byte("To:\r\n\tsomebody\r\n"),
		[]byte("From: \r\n someone\r\n"),
		[]byte("\r\n"),
	}, header.lines)
}

func TestParseHeaderMultilineFilename(t *testing.T) {
	const literal = "Content-Type: application/msword; name=\"this is a very long\nfilename.doc\""

	header, err := ParseHeader([]byte(literal))
	require.NoError(t, err)

	assert.Equal(t, [][]byte{
		[]byte("Content-Type: application/msword; name=\"this is a very long\nfilename.doc\""),
	}, header.lines)
}

func TestParseHeaderMultilineFilenameWithColon(t *testing.T) {
	const literal = "Content-Type: application/msword; name=\"this is a very long\nfilename: too long.doc\""

	header, err := ParseHeader([]byte(literal))
	require.NoError(t, err)

	assert.Equal(t, [][]byte{
		[]byte("Content-Type: application/msword; name=\"this is a very long\nfilename: too long.doc\""),
	}, header.lines)
}

func TestParseHeaderMultilineFilenameWithColonAndNewline(t *testing.T) {
	const literal = "Content-Type: application/msword; name=\"this is a very long\nfilename: too long.doc\"\n"

	header, err := ParseHeader([]byte(literal))
	require.NoError(t, err)

	assert.Equal(t, [][]byte{
		[]byte("Content-Type: application/msword; name=\"this is a very long\nfilename: too long.doc\"\n"),
	}, header.lines)
}

func TestParseHeaderMultilineIndent(t *testing.T) {
	const literal = "Subject: a very\r\n\tlong: line with a colon and indent\r\n \r\nand space line\r\nFrom: sender\r\n"

	header, err := ParseHeader([]byte(literal))
	require.NoError(t, err)

	assert.Equal(t, [][]byte{
		[]byte("Subject: a very\r\n\tlong: line with a colon and indent\r\n \r\nand space line\r\n"),
		[]byte("From: sender\r\n"),
	}, header.lines)
}

func TestParseHeaderMultipleMultilineFilenames(t *testing.T) {
	const literal = `Content-Type: application/msword; name="=E5=B8=B6=E6=9C=89=E5=A4=96=E5=9C=8B=E5=AD=97=E7=AC=A6=E7=9A=84=E9=99=84=E4=
=BB=B6.DOC"
Content-Transfer-Encoding: base64
Content-Disposition: attachment; filename="=E5=B8=B6=E6=9C=89=E5=A4=96=E5=9C=8B=E5=AD=97=E7=AC=A6=E7=9A=84=E9=99=84=E4=
=BB=B6.DOC"
Content-ID: <>
`

	header, err := ParseHeader([]byte(literal))
	require.NoError(t, err)

	assert.Equal(t, [][]byte{
		[]byte("Content-Type: application/msword; name=\"=E5=B8=B6=E6=9C=89=E5=A4=96=E5=9C=8B=E5=AD=97=E7=AC=A6=E7=9A=84=E9=99=84=E4=\n=BB=B6.DOC\"\n"),
		[]byte("Content-Transfer-Encoding: base64\n"),
		[]byte("Content-Disposition: attachment; filename=\"=E5=B8=B6=E6=9C=89=E5=A4=96=E5=9C=8B=E5=AD=97=E7=AC=A6=E7=9A=84=E9=99=84=E4=\n=BB=B6.DOC\"\n"),
		[]byte("Content-ID: <>\n"),
	}, header.lines)
}

func TestSplitHeaderBody(t *testing.T) {
	const literal = "To: user@pm.me\r\n\r\nhi\r\n"

	header, body, err := Split([]byte(literal))
	require.NoError(t, err)

	assert.Equal(t, []byte("To: user@pm.me\r\n\r\n"), header)
	assert.Equal(t, []byte("hi\r\n"), body)
}

func TestSplitHeaderBodyNoBody(t *testing.T) {
	const literal = "To: user@pm.me\r\n\r\n"

	header, body, err := Split([]byte(literal))
	require.NoError(t, err)

	assert.Equal(t, []byte("To: user@pm.me\r\n\r\n"), header)
	assert.Equal(t, []byte(""), body)
}

func TestSplitHeaderBodyOnlyHeader(t *testing.T) {
	const literal = "To: user@pm.me\r\n"

	header, body, err := Split([]byte(literal))
	require.NoError(t, err)

	assert.Equal(t, []byte("To: user@pm.me\r\n"), header)
	assert.Equal(t, []byte(""), body)
}

func TestSplitHeaderBodyOnlyHeaderNoNewline(t *testing.T) {
	const literal = "To: user@pm.me"

	header, body, err := Split([]byte(literal))
	require.NoError(t, err)

	assert.Equal(t, []byte("To: user@pm.me"), header)
	assert.Equal(t, []byte(""), body)
}
