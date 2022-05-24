package tests

import (
	"testing"
)

func TestMessageCreatedUpdate(t *testing.T) {
	runOneToOneTestWithAuth(t, "user", "pass", "/", func(c *testConnection, s *testSession) {
		// Create a mailbox for the test to run in.
		mboxID := s.mailboxCreated("user", []string{"mbox"})

		// Select in the mailbox to receive EXISTS and RECENT updates.
		c.C("A006 select mbox")
		c.Se("A006 OK [READ-WRITE] (^_^)")

		// Start idling in INBOX to receive the updates.
		c.C("A007 IDLE")
		c.S("+ (*_*)")

		// Create some messages externally.
		s.messageCreatedFromFile("user", mboxID, "testdata/multipart-mixed.eml")
		s.messageCreatedFromFile("user", mboxID, "testdata/afternoon-meeting.eml")

		// Expect to receive the updates.
		c.S("* 1 EXISTS", "* 1 RECENT", "* 2 EXISTS", "* 2 RECENT")

		// Stop idling.
		c.C("DONE")
		c.S("A007 OK (^_^)")

		// Create a third message externally.
		s.messageCreatedFromFile("user", mboxID, "testdata/text-plain.eml")

		// Do noop to receive the final updates.
		c.C("A007 NOOP")
		c.S("* 3 EXISTS")
		c.S("* 3 RECENT")
		c.S("A007 OK (^_^)")
	})
}

func TestMessageCreatedNoopUpdate(t *testing.T) {
	runOneToOneTestWithAuth(t, "user", "pass", "/", func(c *testConnection, s *testSession) {
		// Create a mailbox for the test to run in.
		other := s.mailboxCreated("user", []string{"other"})

		// Select in the mailbox to receive EXISTS and RECENT updates.
		c.C(`A001 select other`).OK(`A001`)

		// Create two messages externally.
		s.messageCreatedFromFile("user", other, "testdata/multipart-mixed.eml")
		s.messageCreatedFromFile("user", other, "testdata/afternoon-meeting.eml")

		// Do noop to receive the updates.
		c.C(`A002 NOOP`).Se(`* 2 EXISTS`, `* 2 RECENT`).OK(`A002`)

		// Create two more messages externally.
		s.messageCreatedFromFile("user", other, "testdata/multipart-mixed.eml")
		s.messageCreatedFromFile("user", other, "testdata/afternoon-meeting.eml")

		// Do noop to receive the updates.
		c.C(`A003 NOOP`).Se(`* 4 EXISTS`, `* 4 RECENT`).OK(`A003`)

		// Select away and back to reset the RECENT count.
		c.C(`A004 select inbox`).OK(`A004`)
		c.C(`A005 select other`).OK(`A005`)

		// Create two more messages externally.
		s.messageCreatedFromFile("user", other, "testdata/multipart-mixed.eml")
		s.messageCreatedFromFile("user", other, "testdata/afternoon-meeting.eml")

		// Do noop to receive the updates.
		c.C(`A006 NOOP`).Se(`* 6 EXISTS`, `* 2 RECENT`).OK(`A006`)
	})
}

func TestMessageCreatedIDLEUpdate(t *testing.T) {
	runOneToOneTestWithAuth(t, "user", "pass", "/", func(c *testConnection, s *testSession) {
		// Create a mailbox for the test to run in.
		other := s.mailboxCreated("user", []string{"other"})

		// Select in the mailbox to receive EXISTS and RECENT updates.
		c.C(`A001 select other`).OK(`A001`)

		// Begin IDLE.
		c.C(`A002 IDLE`).S(`+ (*_*)`)

		// Create two messages externally.
		s.messageCreatedFromFile("user", other, "testdata/multipart-mixed.eml")
		s.messageCreatedFromFile("user", other, "testdata/afternoon-meeting.eml")

		// Expect that we receive IDLE updates.
		c.S(`* 1 EXISTS`, `* 1 RECENT`, `* 2 EXISTS`, `* 2 RECENT`)

		// Create two more messages externally.
		s.messageCreatedFromFile("user", other, "testdata/multipart-mixed.eml")
		s.messageCreatedFromFile("user", other, "testdata/afternoon-meeting.eml")

		// Expect that we receive IDLE updates.
		c.S(`* 3 EXISTS`, `* 3 RECENT`, `* 4 EXISTS`, `* 4 RECENT`)

		// Select away and back to reset the RECENT count.
		c.C(`DONE`).OK(`A002`)
		c.C(`A003 select inbox`).OK(`A003`)
		c.C(`A004 select other`).OK(`A004`)
		c.C(`A005 IDLE`).S(`+ (*_*)`)

		// Create two more messages externally.
		s.messageCreatedFromFile("user", other, "testdata/multipart-mixed.eml")
		s.messageCreatedFromFile("user", other, "testdata/afternoon-meeting.eml")

		// Expect that we receive IDLE updates.
		c.S(`* 5 EXISTS`, `* 1 RECENT`, `* 6 EXISTS`, `* 2 RECENT`)

		// Stop IDLE.
		c.C(`DONE`).OK(`A005`)
	})
}

func TestMessageRemovedUpdate(t *testing.T) {
	runOneToOneTestWithAuth(t, "user", "pass", "/", func(c *testConnection, s *testSession) {
		// Create a mailbox for the test to run in.
		mboxID := s.mailboxCreated("user", []string{"mbox"})

		// Create some messages externally.
		messageID1 := s.messageCreatedFromFile("user", mboxID, "testdata/multipart-mixed.eml")
		messageID2 := s.messageCreatedFromFile("user", mboxID, "testdata/afternoon-meeting.eml")
		messageID3 := s.messageCreatedFromFile("user", mboxID, "testdata/text-plain.eml")
		messageID4 := s.messageCreatedFromFile("user", mboxID, "testdata/text-plain.eml")

		// Select in the mailbox to receive EXPUNGE updates.
		c.C("A006 select mbox")
		c.Se("A006 OK [READ-WRITE] (^_^)")

		// Start idling in INBOX to receive the EXPUNGE updates.
		c.C("A007 IDLE")
		c.S("+ (*_*)")

		// Remove the first message.
		s.messageRemoved("user", messageID1, mboxID)

		// Expect to receive the EXPUNGE update of the removed message.
		c.S("* 1 EXPUNGE")

		// Remove the second message.
		s.messageRemoved("user", messageID2, mboxID)

		// Expect to receive the EXPUNGE update of the removed message.
		// Due to the previous EXPUNGE, this message now has sequence number 1.
		c.S("* 1 EXPUNGE")

		// Remove the third message.
		s.messageRemoved("user", messageID3, mboxID)

		// Expect to receive the EXPUNGE update of the removed message.
		// Due to the previous two EXPUNGEs, this message now has sequence number 1.
		c.S("* 1 EXPUNGE")

		// Stop idling.
		c.C("DONE")
		c.S("A007 OK (^_^)")

		// Remove the fourth message.
		s.messageRemoved("user", messageID4, mboxID)

		// The message has been expunged, but we haven't received the EXPUNGE update for it yet.
		// Therefore, we still think there is one message in the mailbox!
		c.C("A007 FETCH 1:* (UID)")
		c.S("* 1 FETCH (UID 4)")
		c.Sx("A007 OK .* command completed in .*")

		// Processing the previous fetch shouldn't have led to an EXPUNGE;
		// a subsequent fetch will return the same result!
		c.C("A007 FETCH 1:* (UID)")
		c.S("* 1 FETCH (UID 4)")
		c.Sx("A007 OK .* command completed in .*")

		// Do NOOP to finally receive the EXPUNGE update.
		c.C("A007 NOOP")
		c.S("* 1 EXPUNGE")
		c.S("A007 OK (^_^)")
	})
}

func TestMessageRemovedUpdateRepeated(t *testing.T) {
	runOneToOneTestWithAuth(t, "user", "pass", "/", func(c *testConnection, s *testSession) {
		// Create a mailbox for the test to run in.
		mboxID := s.mailboxCreated("user", []string{"mbox"})

		for i := 1; i <= 1000; i++ {
			messageID := s.messageCreatedFromFile("user", mboxID, "testdata/multipart-mixed.eml")

			c.C("A006 select mbox")
			c.Se("A006 OK [READ-WRITE] (^_^)")

			c.C("A007 IDLE")
			c.S("+ (*_*)")

			s.messageRemoved("user", messageID, mboxID)
			c.S("* 1 EXPUNGE")

			c.C("DONE")
			c.S("A007 OK (^_^)")
		}
	})
}

func TestMailboxCreatedUpdate(t *testing.T) {
	runOneToOneTestWithAuth(t, "user", "pass", "/", func(c *testConnection, s *testSession) {
		c.C(`A82 LIST "" *`)
		c.S(`* LIST (\Unmarked) "/" "INBOX"`)
		c.S(`A82 OK (^_^)`)

		s.mailboxCreated("user", []string{"some-mailbox"})

		c.C(`A82 LIST "" *`)
		c.S(`* LIST (\Unmarked) "/" "some-mailbox"`,
			`* LIST (\Unmarked) "/" "INBOX"`)
		c.S(`A82 OK (^_^)`)
	})
}

func TestMessageSeenUpdate(t *testing.T) {
	runOneToOneTestWithAuth(t, "user", "pass", "/", func(c *testConnection, s *testSession) {
		mailboxID := s.mailboxCreated("user", []string{"mbox"})
		messageID := s.messageCreatedFromFile("user", mailboxID, "testdata/multipart-mixed.eml")

		c.C("A001 SELECT mbox").OK("A001")

		c.C("A002 FETCH 1 (FLAGS)")
		c.S(`* 1 FETCH (FLAGS (\Recent)`)
		c.OK("A002")

		s.messageSeen("user", messageID, true)

		c.C("A003 FETCH 1 (FLAGS)")
		c.S(`* 1 FETCH (FLAGS (\Recent \Seen)`)
		c.OK("A003")

		s.messageSeen("user", messageID, false)

		c.C("A004 FETCH 1 (FLAGS)")
		c.S(`* 1 FETCH (FLAGS (\Recent)`)
		c.OK("A004")
	})
}

func TestMessageFlaggedUpdate(t *testing.T) {
	runOneToOneTestWithAuth(t, "user", "pass", "/", func(c *testConnection, s *testSession) {
		mailboxID := s.mailboxCreated("user", []string{"mbox"})
		messageID := s.messageCreatedFromFile("user", mailboxID, "testdata/multipart-mixed.eml")

		c.C("A001 SELECT mbox").OK("A001")

		s.messageFlagged("user", messageID, true)
		c.C("A003 FETCH 1 (FLAGS)")
		c.S(`* 1 FETCH (FLAGS (\Flagged \Recent)`)
		c.OK("A003")

		s.messageFlagged("user", messageID, false)
		c.C("A004 FETCH 1 (FLAGS)")
		c.S(`* 1 FETCH (FLAGS (\Recent)`)
		c.OK("A004")
	})
}
