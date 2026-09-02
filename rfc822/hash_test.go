package rfc822

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetMessageHashSameBodyDifferentTextEncodings(t *testing.T) {
	data1, err := os.ReadFile("testdata/hash_quoted.eml")
	require.NoError(t, err)

	data2, err := os.ReadFile("testdata/hash_utf8.eml")
	require.NoError(t, err)

	h1, err := GetMessageHash(data1)
	require.NoError(t, err)

	h2, err := GetMessageHash(data2)
	require.NoError(t, err)

	require.Equal(t, h1, h2)
}

func TestLiteralsStructurallyEquivalentBoundaryDrift(t *testing.T) {
	// multipart/mixed > multipart/related(text/plain + inline image) + attachment
	litA := buildStructuralTestLiteral("outerA", "innerA", "attachment")
	litB := buildStructuralTestLiteral("outerB", "innerB", "attachment")

	require.False(t, bytes.Equal(litA, litB), "raw literals should differ due to boundary tokens")

	eq, err := LiteralsStructurallyEqual(litA, litB)
	require.NoError(t, err)
	require.True(t, eq, "boundary-only drift should be structurally equivalent")

	// Both literals must expose the same MIME tree regardless of boundary values.
	for _, lit := range [][]byte{litA, litB} {
		expectPart(t, lit, MultipartRelated, "", 1)
		expectPart(t, lit, TextPlain, "body", 1, 1)
		expectPart(t, lit, "image/png", "inline", 1, 2)
		expectPart(t, lit, "image/png", "attachment", 2)
	}

	// A real content change (different attachment body) must not be equivalent.
	litC := buildStructuralTestLiteral("outerA", "innerA", "different-attachment")

	eq, err = LiteralsStructurallyEqual(litA, litC)
	require.NoError(t, err)
	require.False(t, eq, "different leaf body should not be structurally equivalent")
}

func TestLiteralsHeadersEqualBoundaryDrift(t *testing.T) {
	litA := buildStructuralTestLiteral("outerA", "innerA", "attachment")
	litB := buildStructuralTestLiteral("outerB", "innerB", "attachment")

	require.False(t, bytes.Equal(litA, litB))

	eq, err := LiteralsHeadersEqual(litA, litB)
	require.NoError(t, err)
	require.True(t, eq, "boundary-only drift should keep the same header fingerprint")
}

func TestLiteralsHeadersEqualHeaderEnrichment(t *testing.T) {
	litA := buildStructuralTestLiteral("outerA", "innerA", "attachment")
	// Header order is irrelevant to the fingerprint, so prepending is fine.
	litEnriched := append([]byte("Message-Id: <id@pm.me>\r\nX-Pm-Internal-Id: abc123\r\n"), litA...)

	structEq, err := LiteralsStructurallyEqual(litA, litEnriched)
	require.NoError(t, err)
	require.True(t, structEq, "content is unchanged, so structure must match")

	headersEq, err := LiteralsHeadersEqual(litA, litEnriched)
	require.NoError(t, err)
	require.False(t, headersEq, "added Message-Id / X-Pm-* headers must change the fingerprint")

	cosmeticEq, err := CompareLiterals(litA, litEnriched)
	require.NoError(t, err)
	require.False(t, cosmeticEq, "header enrichment must not be treated as cosmetically equal")
}

func TestLiteralsHeadersEqualIgnoresGluonID(t *testing.T) {
	litA := buildStructuralTestLiteral("outerA", "innerA", "attachment")

	withGluonID, err := SetHeaderValue(litA, gluonInternalHeaderKey, "some-internal-id")
	require.NoError(t, err)

	require.False(t, bytes.Equal(litA, withGluonID))

	eq, err := LiteralsHeadersEqual(litA, withGluonID)
	require.NoError(t, err)
	require.True(t, eq, "X-Pm-Gluon-Id must be ignored by the header fingerprint")
}

func TestCompareLiterals(t *testing.T) {
	litA := buildStructuralTestLiteral("outerA", "innerA", "attachment")
	litB := buildStructuralTestLiteral("outerB", "innerB", "attachment")
	litC := buildStructuralTestLiteral("outerA", "innerA", "different-attachment")

	eq, err := CompareLiterals(litA, litB)
	require.NoError(t, err)
	require.True(t, eq, "boundary-only drift is cosmetically equal")

	eq, err = CompareLiterals(litA, litC)
	require.NoError(t, err)
	require.False(t, eq, "different content is not cosmetically equal")
}

func buildStructuralTestLiteral(outerBoundary, innerBoundary, attachmentBody string) []byte {
	return []byte("From: sender@example.com\r\n" +
		"To: recipient@example.com\r\n" +
		"Subject: structural test\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: multipart/mixed; boundary=\"" + outerBoundary + "\"\r\n" +
		"\r\n" +
		"--" + outerBoundary + "\r\n" +
		"Content-Type: multipart/related; boundary=\"" + innerBoundary + "\"\r\n" +
		"\r\n" +
		"--" + innerBoundary + "\r\n" +
		"Content-Type: text/plain\r\n" +
		"\r\n" +
		"body\r\n" +
		"--" + innerBoundary + "\r\n" +
		"Content-Type: image/png\r\n" +
		"Content-Disposition: inline\r\n" +
		"\r\n" +
		"inline\r\n" +
		"--" + innerBoundary + "--\r\n" +
		"--" + outerBoundary + "\r\n" +
		"Content-Type: image/png\r\n" +
		"Content-Disposition: attachment\r\n" +
		"\r\n" +
		attachmentBody + "\r\n" +
		"--" + outerBoundary + "--\r\n")
}

// expectPart asserts the Content-Type (and, for leaf parts, the trimmed body) of the MIME part at the given 1-based path.
func expectPart(t *testing.T, literal []byte, wantType MIMEType, wantBody string, path ...int) {
	t.Helper()

	part, err := Parse(literal).Part(path...)
	require.NoError(t, err, "part %v should exist", path)

	mimeType, _, err := part.ContentType()
	require.NoError(t, err)
	require.Equal(t, wantType, mimeType, "content type of part %v", path)

	if wantBody != "" {
		require.Equal(t, wantBody, string(bytes.TrimSpace(part.Body())), "body of part %v", path)
	}
}
