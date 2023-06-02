package tests

import (
	"regexp"
	"testing"
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/stretchr/testify/require"
)

func TestDraftScenario(t *testing.T) {
	// Simulate a draft update issued from the connector, which involves deleting the original message in drafts
	// and replacing it with a new one.
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		mailboxID := s.mailboxCreated("user", []string{"Drafts"})

		c.C("A002 SELECT Drafts").OK("A002")

		messageID := s.messageCreated("user", mailboxID, []byte("To: 3@3.pm"), time.Now())

		c.C("A002 NOOP")
		c.S("* 1 EXISTS")
		c.S("* 1 RECENT")
		c.OK("A002")

		c.C("A003 FETCH 1 (BODY.PEEK[HEADER.FIELDS (To)])")
		c.S("* 1 FETCH (BODY[HEADER.FIELDS (TO)] {10}\r\nTo: 3@3.pm)")
		c.OK("A003")

		s.messageDeleted("user", messageID)
		s.messageCreated("user", mailboxID, []byte("To: 4@4.pm"), time.Now())
		s.flush("user")

		c.C("A004 NOOP")
		c.S("* 1 EXPUNGE")
		c.S("* 1 EXISTS")
		c.S("* 1 RECENT")
		c.OK("A004")

		c.C("A005 FETCH 1 (BODY.PEEK[HEADER.FIELDS (To)])")
		c.S("* 1 FETCH (BODY[HEADER.FIELDS (TO)] {10}\r\nTo: 4@4.pm)")
		c.OK("A005")
	})
}

func TestDraftSavedAgain(t *testing.T) {
	// Email client can save same message i.e.: it will delete old draft
	// and append one with same content.
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		mailboxID := s.mailboxCreatedWithAttributes("user", []string{"Drafts"}, imap.NewFlagSet(imap.AttrDrafts))

		c.C("A002 SELECT Drafts").OK("A002")

		_ = s.messageCreated("user", mailboxID, []byte(buildRFC5322TestLiteral("To: 3@3.pm")), time.Now())
		s.flush("user")

		c.C("A002 NOOP")
		c.S("* 1 EXISTS")
		c.S("* 1 RECENT")
		c.OK("A002")

		c.C("A003 FETCH 1 (BODY.PEEK[])")
		raw := c.read()

		fetch := regexp.MustCompile(
			regexp.QuoteMeta("* 1 FETCH (BODY[] {134}\r\nX-Pm-Gluon-Id: ") +
				".*" +
				regexp.QuoteMeta("\r\n"+buildRFC5322TestLiteral("To: 3@3.pm")+")\r\n"),
		)

		require.Regexp(t, fetch, string(raw))
		c.OK("A003")

		// Expunge+Append same message
		{
			c.C(`A004 STORE 1 +FLAGS (\Deleted)`).OK("A004")
			c.C(`A005 EXPUNGE`).OK("A005")

			c.doAppend(
				"Drafts",
				string(raw[25:len(raw)-3]),
			).expect("OK")

			c.C("A004 FETCH 1 (BODY.PEEK[])")
			raw := c.read()
			require.Regexp(t, fetch, string(raw))
			c.OK("A004")
		}

		newBody2 := string(raw[25:]) + "\r\nThis is body\r\n"
		fetchUpdated2 := regexp.MustCompile(
			regexp.QuoteMeta("* 1 FETCH (BODY[] {153}\r\nX-Pm-Gluon-Id: ") +
				".*" +
				regexp.QuoteMeta("\r\n"+buildRFC5322TestLiteral("To: 3@3.pm)")+"\r\n\r\nThis is body\r\n)"),
		)
		// Expunge+Append different message
		{
			c.C(`A004 STORE 1 +FLAGS (\Deleted)`).OK("A004")
			c.C(`A005 EXPUNGE`).OK("A005")

			c.doAppend(
				"Drafts",
				newBody2,
			).expect("OK")

			c.C("A004 FETCH 1 (BODY.PEEK[])")
			raw := c.read()
			require.Regexp(t, fetchUpdated2, string(raw))
			c.OK("A004")
		}

		// Append + Expunge different message
		newBody3 := newBody2 + "hello\r\n"
		fetchUpdated3 := regexp.MustCompile(
			regexp.QuoteMeta("* ") +
				"\\d" +
				regexp.QuoteMeta(" FETCH (BODY[] {160}\r\nX-Pm-Gluon-Id: ") +
				".*" +
				regexp.QuoteMeta("\r\n"+buildRFC5322TestLiteral("To: 3@3.pm)")+"\r\n\r\nThis is body\r\nhello\r\n)"),
		)
		{

			c.doAppend(
				"Drafts",
				newBody3,
			).expect("OK")

			c.C("A004 FETCH 1 (BODY.PEEK[])")
			raw := c.read()
			require.Regexp(t, fetchUpdated2, string(raw))
			c.OK("A004")

			c.C("A005 FETCH 2 (BODY.PEEK[])")
			raw = c.read()
			require.Regexp(t, fetchUpdated3, string(raw))
			c.OK("A005")

			c.C(`A004 STORE 1 +FLAGS (\Deleted)`).OK("A004")
			c.C(`A005 EXPUNGE`).OK("A005")

			c.C("A006 FETCH 1 (BODY.PEEK[])")
			raw = c.read()
			require.Regexp(t, fetchUpdated3, string(raw))
			c.OK("A006")

		}
	})
}
