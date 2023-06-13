package tests

import (
	"testing"
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/utils"
	"github.com/emersion/go-imap/client"
	"github.com/stretchr/testify/require"
)

func TestMessageCreatedUpdate(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		// Create a mailbox for the test to run in.
		mboxID := s.mailboxCreated("user", []string{"mbox"})

		// Select in the mailbox to receive EXISTS and RECENT updates.
		c.C("A006 select mbox")
		c.Se("A006 OK [READ-WRITE] SELECT")

		// Start idling in INBOX to receive the updates.
		c.C("A007 IDLE")
		c.S("+ Ready")

		// Create some messages externally.
		s.messageCreatedFromFile("user", mboxID, "testdata/multipart-mixed.eml")
		s.messageCreatedFromFile("user", mboxID, "testdata/afternoon-meeting.eml")

		// Expect to receive the updates.
		c.S("* 2 EXISTS", "* 2 RECENT")

		// Stop idling.
		c.C("DONE")
		c.OK("A007")

		// Create a third message externally.
		s.messageCreatedFromFile("user", mboxID, "testdata/text-plain.eml")

		// Do noop to receive the final updates.
		c.C("A007 NOOP")
		c.S("* 3 EXISTS")
		c.S("* 3 RECENT")
		c.OK("A007")
	})
}

func TestMessageCreatedNoopUpdate(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		// Create a mailbox for the test to run in.
		other := s.mailboxCreated("user", []string{"other"})

		// Select in the mailbox to receive EXISTS and RECENT updates.
		c.C(`A001 select other`).OK(`A001`)

		// Create two messages externally.
		s.messageCreatedFromFile("user", other, "testdata/multipart-mixed.eml")
		s.messageCreatedFromFile("user", other, "testdata/afternoon-meeting.eml")
		s.flush("user")

		// Do noop to receive the updates.
		c.C(`A002 NOOP`).Se(`* 2 EXISTS`, `* 2 RECENT`).OK(`A002`)

		// Create two more messages externally.
		s.messageCreatedFromFile("user", other, "testdata/multipart-mixed.eml")
		s.messageCreatedFromFile("user", other, "testdata/afternoon-meeting.eml")
		s.flush("user")

		// Do noop to receive the updates.
		c.C(`A003 NOOP`).Se(`* 4 EXISTS`, `* 4 RECENT`).OK(`A003`)

		// Select away and back to reset the RECENT count.
		c.C(`A004 select inbox`).OK(`A004`)
		c.C(`A005 select other`).OK(`A005`)

		// Create two more messages externally.
		s.messageCreatedFromFile("user", other, "testdata/multipart-mixed.eml")
		s.messageCreatedFromFile("user", other, "testdata/afternoon-meeting.eml")
		s.flush("user")

		// Do noop to receive the updates.
		c.C(`A006 NOOP`).Se(`* 6 EXISTS`, `* 2 RECENT`).OK(`A006`)
	})
}

func TestMessageCreatedIDLEUpdate(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		// Create a mailbox for the test to run in.
		other := s.mailboxCreated("user", []string{"other"})

		// Select in the mailbox to receive EXISTS and RECENT updates.
		c.C(`A001 select other`).OK(`A001`)

		// Begin IDLE.
		c.C(`A002 IDLE`).S(`+ Ready`)

		// Create two messages externally.
		s.messageCreatedFromFile("user", other, "testdata/multipart-mixed.eml")
		s.messageCreatedFromFile("user", other, "testdata/afternoon-meeting.eml")
		s.flush("user")

		// Expect that we receive IDLE updates.
		c.S(`* 2 EXISTS`, `* 2 RECENT`)

		// Create two more messages externally.
		s.messageCreatedFromFile("user", other, "testdata/multipart-mixed.eml")
		s.messageCreatedFromFile("user", other, "testdata/afternoon-meeting.eml")
		s.flush("user")

		// Expect that we receive IDLE updates.
		c.S(`* 4 EXISTS`, `* 4 RECENT`)

		// Select away and back to reset the RECENT count.
		c.C(`DONE`).OK(`A002`)
		c.C(`A003 select inbox`).OK(`A003`)
		c.C(`A004 select other`).OK(`A004`)
		c.C(`A005 IDLE`).S(`+ Ready`)

		// Create two more messages externally.
		s.messageCreatedFromFile("user", other, "testdata/multipart-mixed.eml")
		s.messageCreatedFromFile("user", other, "testdata/afternoon-meeting.eml")
		s.flush("user")

		// Expect that we receive IDLE updates.
		c.S(`* 6 EXISTS`, `* 2 RECENT`)

		// Stop IDLE.
		c.C(`DONE`).OK(`A005`)
	})
}

func TestMessageRemovedUpdate(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		// Create a mailbox for the test to run in.
		mboxID := s.mailboxCreated("user", []string{"mbox"})

		// Create some messages externally.
		messageID1 := s.messageCreatedFromFile("user", mboxID, "testdata/multipart-mixed.eml")
		messageID2 := s.messageCreatedFromFile("user", mboxID, "testdata/afternoon-meeting.eml")
		messageID3 := s.messageCreatedFromFile("user", mboxID, "testdata/text-plain.eml")
		messageID4 := s.messageCreatedFromFile("user", mboxID, "testdata/text-plain.eml")

		// Select in the mailbox to receive EXPUNGE updates.
		c.C("A006 select mbox")
		c.Se("A006 OK [READ-WRITE] SELECT")

		// Start idling in INBOX to receive the EXPUNGE updates.
		c.C("A007 IDLE")
		c.S("+ Ready")

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
		c.OK("A007")

		// Remove the fourth message.
		s.messageRemoved("user", messageID4, mboxID)

		// The message has been expunged, but we haven't received the EXPUNGE update for it yet.
		// Therefore, we still think there is one message in the mailbox!
		c.C("A007 FETCH 1:* (UID)")
		c.S("* 1 FETCH (UID 4)")
		c.OK("A007")

		// Processing the previous fetch shouldn't have led to an EXPUNGE;
		// a subsequent fetch will return the same result!
		c.C("A007 FETCH 1:* (UID)")
		c.S("* 1 FETCH (UID 4)")
		c.OK("A007")

		// Do NOOP to finally receive the EXPUNGE update.
		c.C("A007 NOOP")
		c.S("* 1 EXPUNGE")
		c.OK("A007")
	})
}

func TestMessageRemovedUpdateRepeated(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t, withIdleBulkTime(0)), func(c *testConnection, s *testSession) {
		// Create a mailbox for the test to run in.
		mboxID := s.mailboxCreated("user", []string{"mbox"})

		for i := 1; i <= 1000; i++ {
			messageID := s.messageCreatedFromFile("user", mboxID, "testdata/multipart-mixed.eml")

			c.C("A006 select mbox")
			c.Se("A006 OK [READ-WRITE] SELECT")

			c.C("A007 IDLE")
			c.S("+ Ready")

			s.messageRemoved("user", messageID, mboxID)
			c.S("* 1 EXPUNGE")

			c.C("DONE")
			c.OK("A007")
		}
	})
}

func TestMailboxCreatedUpdate(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		c.C(`A82 LIST "" *`)
		c.S(`* LIST (\Unmarked) "/" "INBOX"`)
		c.OK("A82")

		s.mailboxCreated("user", []string{"some-mailbox"})

		c.C(`A82 LIST "" *`)
		c.S(`* LIST (\Unmarked) "/" "some-mailbox"`,
			`* LIST (\Unmarked) "/" "INBOX"`)
		c.OK("A82")
	})
}

func TestMessageSeenUpdate(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		mailboxID := s.mailboxCreated("user", []string{"mbox"})
		messageID := s.messageCreatedFromFile("user", mailboxID, "testdata/multipart-mixed.eml")

		c.C("A001 SELECT mbox").OK("A001")

		c.C("A002 FETCH 1 (FLAGS)")
		c.S(`* 1 FETCH (FLAGS (\Recent))`)
		c.OK("A002")

		s.messageSeen("user", messageID, true)

		c.C("A003 FETCH 1 (FLAGS)")
		c.S(`* 1 FETCH (FLAGS (\Recent))`)
		// Unilateral update arrives after fetch.
		c.S(`* 1 FETCH (FLAGS (\Recent \Seen))`)
		c.OK("A003")

		s.messageSeen("user", messageID, false)

		c.C("A004 FETCH 1 (FLAGS)")
		c.S(`* 1 FETCH (FLAGS (\Recent \Seen))`)
		// Unilateral update arrives after fetch.
		c.S(`* 1 FETCH (FLAGS (\Recent))`)
		c.OK("A004")
	})
}

func TestMessageFlaggedUpdate(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		mailboxID := s.mailboxCreated("user", []string{"mbox"})
		messageID := s.messageCreatedFromFile("user", mailboxID, "testdata/multipart-mixed.eml")

		c.C("A001 SELECT mbox").OK("A001")

		s.messageFlagged("user", messageID, true)
		c.C("A003 FETCH 1 (FLAGS)")
		c.S(`* 1 FETCH (FLAGS (\Recent))`)
		c.OK("A003")

		s.messageFlagged("user", messageID, false)
		c.C("A004 FETCH 1 (FLAGS)")
		c.S(`* 1 FETCH (FLAGS (\Flagged \Recent))`)
		// Unilateral updates arrive afterwards.
		c.S(`* 1 FETCH (FLAGS (\Recent))`)
		c.OK("A004")
	})
}

func TestMessageAddWithSameID(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		mailboxID := s.mailboxCreated("user", []string{"mbox"})
		flags := []string{imap.FlagFlagged, imap.FlagDraft, "\\foo", "\\bar", imap.AttrMarked}
		messageID := imap.MessageID(utils.NewRandomMessageID())
		s.batchMessageCreatedWithID("user", mailboxID, 2, func(i int) (imap.MessageID, []byte, []string) {
			return messageID, []byte("to: 1@1.com"), flags
		})

		s.flush("user")

		c.C("A001 SELECT mbox").OK("A001")

		c.C("A003 FETCH 1 (FLAGS)")
		c.S(`* 1 FETCH (FLAGS (\Draft \Flagged \Marked \Recent \bar \foo))`)
		c.OK("A003")
	})
}

func TestBatchMessageAddedWithMultipleFlags(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		mailboxID := s.mailboxCreated("user", []string{"mbox"})
		flags := []string{imap.FlagFlagged, imap.FlagDraft, "\\foo", "\\bar", imap.AttrMarked}
		s.batchMessageCreated("user", mailboxID, 2, func(i int) ([]byte, []string) {
			return []byte("to: 1@1.com"), flags
		})

		s.flush("user")
	})
}

func TestMessageCreatedWithIgnoreMissingMailbox(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(c *client.Client, s *testSession) {
		mailboxID := s.mailboxCreated("user", []string{"mbox"})
		s.setUpdatesAllowedToFail("user", true)
		{
			// First round fails as a missing mailbox is not allowed.
			s.messageCreatedWithMailboxes("user", []imap.MailboxID{mailboxID, "THIS MAILBOX DOES NOT EXISTS"}, []byte("To: Test"), time.Now())
			status, err := c.Select("mbox", false)
			require.NoError(t, err)
			require.Equal(t, status.Messages, uint32(0))
		}
		{
			// Second round succeeds as we publish an update that is allowed to fail.
			s.setAllowMessageCreateWithUnknownMailboxID("user", true)
			s.messageCreatedWithMailboxes("user", []imap.MailboxID{mailboxID, "THIS MAILBOX DOES NOT EXISTS"}, []byte("To: Test"), time.Now())
			status, err := c.Select("mbox", false)
			require.NoError(t, err)
			require.Equal(t, status.Messages, uint32(1))
		}
	})
}

func TestInvalidIMAPCommandDoesNotBlockStateUpdates(t *testing.T) {
	// Test that a sequence of delete followed by create with the same message ID  results in an updated message.
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		mailboxID := s.mailboxCreated("user", []string{"mbox1"})
		messageID := s.messageCreated("user", mailboxID, []byte("To: 3@3.pm"), time.Now())

		s.flush("user")
		c.C(`A002 SELECT mbox1`).OK(`A002`)

		// Check that the message exists.
		c.C(`A005 FETCH 1 (BODY[HEADER.FIELDS (TO)])`)
		c.S("* 1 FETCH (BODY[HEADER.FIELDS (TO)] {10}\r\nTo: 3@3.pm FLAGS (\\Recent \\Seen))")
		c.OK("A005")

		// Update to the message is applied.
		s.messageDeleted("user", messageID)
		s.flush("user")

		// First fetch should read the old message data due to deferred deletion.
		c.C(`A005 FETCH 1 (BODY[HEADER.FIELDS (TO)])`)
		c.S("* 1 FETCH (BODY[HEADER.FIELDS (TO)] {10}\r\nTo: 3@3.pm)")
		c.OK("A005")

		c.C("A006 NOOP").OK("")

		// Second fetch fail with
		c.C(`A005 FETCH 1 (BODY[HEADER.FIELDS (TO)])`)
		c.S("A005 BAD no such message")

		// new message gets create
		s.messageCreated("user", mailboxID, []byte("To: 3@3.pm"), time.Now())
		s.flush("user")

		// Third fetch will still fail, but now we get updates to the state.
		c.C(`A005 FETCH 1 (BODY[HEADER.FIELDS (TO)])`)
		c.S(`* 1 EXISTS`)
		c.S(`* 1 RECENT`)
		c.S("A005 BAD no such message")

		// Forth fetch succeeds.
		c.C(`A005 FETCH 1 (BODY[HEADER.FIELDS (TO)])`)
		c.S("* 1 FETCH (BODY[HEADER.FIELDS (TO)] {10}\r\nTo: 3@3.pm FLAGS (\\Recent \\Seen))")
		c.OK("A005")
	})
}

/* GODT-2688 - Investigate test flakiness.
func TestUpdatedMessageFetchSucceedsOnSecondTry(t *testing.T) {
	// Test that a sequence of delete followed by create with the same message ID  results in an updated message.
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		mailboxID := s.mailboxCreated("user", []string{"mbox1"})
		messageID := s.messageCreated("user", mailboxID, []byte("To: 3@3.pm"), time.Now())

		s.flush("user")
		c.C(`A002 SELECT mbox1`).OK(`A002`)

		// Check that the message exists.
		c.C(`A005 FETCH 1 (BODY[HEADER.FIELDS (TO)])`)
		c.S("* 1 FETCH (BODY[HEADER.FIELDS (TO)] {10}\r\nTo: 3@3.pm FLAGS (\\Recent \\Seen))")
		c.OK("A005")

		// Update to the message is applied.
		s.messageUpdatedWithID("user", messageID, mailboxID, []byte("To: 4@4.pm"), time.Now())
		s.flush("user")

		// First fetch should read the old message data due to deferred deletion and get the updates
		// that the message has been removed and replaced with new version.
		c.C(`A005 FETCH 1 (BODY[HEADER.FIELDS (TO)])`)
		c.S("* 1 FETCH (BODY[HEADER.FIELDS (TO)] {10}\r\nTo: 3@3.pm)")
		c.S("* 2 EXISTS")
		c.S("* 2 RECENT")
		c.OK("A005", "EXPUNGEISSUED")

		// Original message remains active until Noop/Status/Select.
		c.C(`A005 FETCH 1 (BODY[HEADER.FIELDS (TO)])`)
		c.S("* 1 FETCH (BODY[HEADER.FIELDS (TO)] {10}\r\nTo: 3@3.pm)")
		c.OK(`A005`)

		// New message is now available.
		c.C(`A005 FETCH 2 (BODY[HEADER.FIELDS (TO)])`)
		c.S("* 2 FETCH (BODY[HEADER.FIELDS (TO)] {10}\r\nTo: 4@4.pm FLAGS (\\Recent \\Seen))")
		c.OK("A005")

		// Noop finally allows expunges.
		c.C("A006 NOOP")
		c.OK("A006")

		// Old message is now gone.
		c.C(`A005 FETCH 1 (BODY[HEADER.FIELDS (TO)])`)
		c.S("* 1 FETCH (BODY[HEADER.FIELDS (TO)] {10}\r\nTo: 4@4.pm)")
		c.OK("A005")
	})
}*/

func TestDeleteMailboxFromConnectorAlsoRemoveSubscriptionStatus(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		mailboxID := s.mailboxCreated("user", []string{"mbox1"})
		s.flush("user")
		s.mailboxDeleted("user", mailboxID)
		s.flush("user")

		c.C(`S001 LSUB "" "*"`)
		c.S(
			`* LSUB (\Unmarked) "/" "INBOX"`,
		)
		c.S("S001 OK LSUB")
	})
}
