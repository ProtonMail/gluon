package tests

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func buildStructuralTestLiteral(outerBoundary, innerBoundary, attachmentBody string) []byte {
	return []byte("From: sender@example.com\r\n" +
		"To: recipient@example.com\r\n" +
		"Subject: structural test\r\n" +
		"Date: " + DefaultTestDateStr + "\r\n" +
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

// TestRemoteMessageUpdateStructurallyEquivalentKeepsMessage verifies that an update whose literal differs only in
// MIME boundary tokens is treated as unchanged.
// No new UID is assigned so UIDNEXT does not advance.
func TestRemoteMessageUpdateStructurallyEquivalentKeepsMessage(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		mailboxID := s.mailboxCreated("user", []string{"mbox1"})
		messageID := s.messageCreated("user", mailboxID, buildStructuralTestLiteral("outerA", "innerA", "attachment"), time.Now())
		s.flush("user")

		c.C(`A001 STATUS mbox1 (UIDNEXT MESSAGES)`)
		c.S(`* STATUS "mbox1" (UIDNEXT 2 MESSAGES 1)`)
		c.OK(`A001`)

		// Same structure and content, only the boundary tokens differ.
		s.messageUpdatedWithID("user", messageID, mailboxID, buildStructuralTestLiteral("outerB", "innerB", "attachment"), time.Now())
		s.flush("user")

		// No new UID: the message stays in place (UIDNEXT unchanged, still 1 message).
		c.C(`A002 STATUS mbox1 (UIDNEXT MESSAGES)`)
		c.S(`* STATUS "mbox1" (UIDNEXT 2 MESSAGES 1)`)
		c.OK(`A002`)

		// The original message (UID 1) is still present and fetchable.
		c.C(`A003 SELECT mbox1`).OK(`A003`)
		c.C(`A004 FETCH 1 (UID)`)
		c.Sx(`\* 1 FETCH \(UID 1[ )]`)
		c.OK(`A004`)
	})
}

// TestRemoteMessageUpdateStructurallyDifferentAppliesNewLiteral verifies that a
// real content change still takes the new-literal path.
// A new UID is assigned so UIDNEXT advances.
func TestRemoteMessageUpdateStructurallyDifferentAppliesNewLiteral(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		mailboxID := s.mailboxCreated("user", []string{"mbox1"})
		messageID := s.messageCreated("user", mailboxID, buildStructuralTestLiteral("outerA", "innerA", "attachment"), time.Now())
		s.flush("user")

		c.C(`A001 STATUS mbox1 (UIDNEXT MESSAGES)`)
		c.S(`* STATUS "mbox1" (UIDNEXT 2 MESSAGES 1)`)
		c.OK(`A001`)

		// Real content change: the attachment body differs.
		s.messageUpdatedWithID("user", messageID, mailboxID, buildStructuralTestLiteral("outerA", "innerA", "different-attachment"), time.Now())
		s.flush("user")

		// A new UID was assigned (UIDNEXT advanced) while the message count stays 1.
		c.C(`A002 STATUS mbox1 (UIDNEXT MESSAGES)`)
		c.S(`* STATUS "mbox1" (UIDNEXT 3 MESSAGES 1)`)
		c.OK(`A002`)

		c.C(`A003 SELECT mbox1`).OK(`A003`)
		c.C(`A004 FETCH 1 (UID)`)
		c.Sx(`\* 1 FETCH \(UID 2[ )]`)
		c.OK(`A004`)
	})
}

// TestRemoteMessageUpdateHeaderEnrichmentAppliesNewLiteral verifies that an update whose content is unchanged but
// whose headers were updated takes the new-literal path so clients
// see the updated headers. A new UID is assigned so UIDNEXT advances.
func TestRemoteMessageUpdateHeaderEnrichmentAppliesNewLiteral(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		mailboxID := s.mailboxCreated("user", []string{"mbox1"})
		base := buildStructuralTestLiteral("outerA", "innerA", "attachment")
		messageID := s.messageCreated("user", mailboxID, base, time.Now())
		s.flush("user")

		c.C(`A001 STATUS mbox1 (UIDNEXT MESSAGES)`)
		c.S(`* STATUS "mbox1" (UIDNEXT 2 MESSAGES 1)`)
		c.OK(`A001`)

		// Same content and boundaries, but the API added Message-Id / X-Pm-* headers.
		enriched := append([]byte("Message-Id: <id@pm.me>\r\nX-Pm-Internal-Id: abc123\r\n"), base...)
		s.messageUpdatedWithID("user", messageID, mailboxID, enriched, time.Now())
		s.flush("user")

		// The added headers are a real change: a new UID was assigned (UIDNEXT advanced).
		c.C(`A002 STATUS mbox1 (UIDNEXT MESSAGES)`)
		c.S(`* STATUS "mbox1" (UIDNEXT 3 MESSAGES 1)`)
		c.OK(`A002`)

		// The updated message (UID 2) exposes the newly added Message-Id.
		c.C(`A003 SELECT mbox1`).OK(`A003`)
		c.C(`A004 FETCH 1 (UID BODY.PEEK[HEADER.FIELDS (Message-Id)])`)
		c.Sx(`\* 1 FETCH \(UID 2 BODY\[HEADER.FIELDS \(MESSAGE-ID\)\] \{.*`)
		c.OK(`A004`)
	})
}

// TestRemoteMessageUpdateStructurallyEquivalentWithMissingLiteral verifies that when the on-disk literal is missing,
// even a structurally-equivalent update falls back to the new-literal path rather than the mailboxes-only path.
// A new UID is assigned so UIDNEXT advances.
func TestRemoteMessageUpdateStructurallyEquivalentWithMissingLiteral(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "test_store")

	runOneToOneTestWithAuth(t, defaultServerOptions(t, withDataDir(dataDir)), func(c *testConnection, s *testSession) {
		mailboxID := s.mailboxCreated("user", []string{"mbox1"})
		messageID := s.messageCreated("user", mailboxID, buildStructuralTestLiteral("outerA", "innerA", "attachment"), time.Now())
		s.flush("user")

		c.C(`A001 STATUS mbox1 (UIDNEXT MESSAGES)`)
		c.S(`* STATUS "mbox1" (UIDNEXT 2 MESSAGES 1)`)
		c.OK(`A001`)

		// Remove the on-disk literals so the structural comparison has nothing to compare against.
		require.NoError(t, os.RemoveAll(dataDir))

		// Boundary-only drift, but the literal is missing: must take the new-literal path.
		s.messageUpdatedWithID("user", messageID, mailboxID, buildStructuralTestLiteral("outerB", "innerB", "attachment"), time.Now())
		s.flush("user")

		c.C(`A002 STATUS mbox1 (UIDNEXT MESSAGES)`)
		c.S(`* STATUS "mbox1" (UIDNEXT 3 MESSAGES 1)`)
		c.OK(`A002`)
	})
}
