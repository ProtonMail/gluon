package tests

import "testing"

func TestDraftScenario(t *testing.T) {
	// Simulate a draft update issued from the connector, which involves deleting the original message in drafts
	// and replacing it with a new one.
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {

		mailboxID := s.mailboxCreated("user", []string{"Drafts"})

		c.C("A002 SELECT Drafts").OK("A002")

		messageID := s.messageCreated("user", mailboxID, []byte("To: 3@3.pm"))

		c.C("A002 NOOP")
		c.S("* 1 EXISTS")
		c.S("* 1 RECENT")
		c.S("A002 OK")

		c.C("A003 FETCH 1 (BODY.PEEK[HEADER.FIELDS (To)])")
		c.S("* 1 FETCH (BODY[HEADER.FIELDS (TO)] {10}\r\nTo: 3@3.pm)")
		c.OK("A003")

		s.messageDeleted("user", messageID)
		s.messageCreated("user", mailboxID, []byte("To: 4@4.pm"))

		c.C("A002 NOOP")
		c.S("* 1 EXPUNGE")
		c.S("* 1 EXISTS")
		c.S("* 1 RECENT")
		c.S("A002 OK")

		c.C("A003 FETCH 1 (BODY.PEEK[HEADER.FIELDS (To)])")
		c.S("* 1 FETCH (BODY[HEADER.FIELDS (TO)] {10}\r\nTo: 4@4.pm)")
		c.OK("A003")
	})
}
