package tests

import (
	"fmt"
	"regexp"
	"testing"
	"time"

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
		mailboxID := s.mailboxCreated("user", []string{"Drafts"})

		c.C("A002 SELECT Drafts").OK("A002")

		_ = s.messageCreated("user", mailboxID, []byte("To: 3@3.pm"), time.Now())

		c.C("A002 NOOP")
		c.S("* 1 EXISTS")
		c.S("* 1 RECENT")
		c.OK("A002")

		c.C("A003 FETCH 1 (BODY.PEEK[])")
		raw := c.read()

		fetch := regexp.MustCompile(
			regexp.QuoteMeta("* 1 FETCH (BODY[] {63}\r\nX-Pm-Gluon-Id: ") +
				".*" +
				regexp.QuoteMeta("\r\nTo: 3@3.pm)\r\n"),
		)

		require.Regexp(t, fetch, string(raw))
		c.OK("A003")

		// Expunge+Append same message
		{
			c.C(`A004 STORE 1 +FLAGS (\Deleted)`).OK("A004")
			c.C(`A005 EXPUNGE`).OK("A005")

			c.doAppend(
				"Drafts",
				string(raw[24:len(raw)-3]),
			).expect("OK")

			c.C("A004 FETCH 1 (BODY.PEEK[])")
			raw := c.read()
			require.Regexp(t, fetch, string(raw))
			c.OK("A004")
		}

		// Expunge+Append different message
		{

			newBody := string(raw[24:]) + "\r\nThis is body\r\n"

			c.C(`A004 STORE 1 +FLAGS (\Deleted)`).OK("A004")
			c.C(`A005 EXPUNGE`).OK("A005")

			c.doAppend(
				"Drafts",
				newBody,
			).expect("OK")

			c.C("A004 FETCH 1 (BODY.PEEK[])")
			c.S(fmt.Sprintf("* 1 FETCH (BODY[] {%d}\r\n%s)", len(newBody), newBody))
			c.OK("A004")
		}

	})
}
