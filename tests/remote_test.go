package tests

import (
	"testing"

	"github.com/ProtonMail/gluon/imap"
)

func TestRemoteCopy(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		s.setFolderPrefix("user", "Folders")
		s.setLabelPrefix("user", "Labels")

		// Create two exclusive mailboxes, one with 100 messages.
		s.mailboxCreated("user", []string{"Folders", "mbox1"}, "testdata/dovecot-crlf")
		s.mailboxCreated("user", []string{"Folders", "mbox2"})

		// Create two non-exclusive mailboxes, one with 100 messages.
		s.mailboxCreated("user", []string{"Labels", "mbox1"}, "testdata/dovecot-crlf")
		s.mailboxCreated("user", []string{"Labels", "mbox2"})

		// Copy everything from Folders/mbox1 to Folders/mbox2.
		c.C(`A001 select Folders/mbox1`).Sxe(`100 EXISTS`).OK(`A001`)
		c.C(`A002 copy 1:* Folders/mbox2`).OK(`A002`)

		// Copy everything from Labels/mbox1 to Labels/mbox2.
		c.C(`A001 select Labels/mbox1`).Sxe(`100 EXISTS`).OK(`A001`)
		c.C(`A002 copy 1:* Labels/mbox2`).OK(`A002`)

		// The folders are exclusive and so the remote will remove them automatically from the origin.
		// The labels are non-exclusive and so they will not be modified by the remote.
		s.flush("user")

		// Check that the message counts are expected.
		c.C(`A003 noop`).OK(`A003`)
		c.C(`A004 status Folders/mbox1 (messages)`).Sx(`MESSAGES 0`).OK(`A004`)
		c.C(`A005 status Folders/mbox2 (messages)`).Sx(`MESSAGES 100`).OK(`A005`)
		c.C(`A004 status Labels/mbox1 (messages)`).Sx(`MESSAGES 100`).OK(`A004`)
		c.C(`A005 status Labels/mbox2 (messages)`).Sx(`MESSAGES 100`).OK(`A005`)
	})
}

func TestRemoteDeletionPool(t *testing.T) {
	runManyToOneTestWithAuth(t, defaultServerOptions(t), []int{1, 2}, func(c map[int]*testConnection, s *testSession) {
		// Create a mailbox for the test to run in.
		mboxID1 := s.mailboxCreated("user", []string{"mbox1"})
		mboxID2 := s.mailboxCreated("user", []string{"mbox2"})

		// Create some messages externally.
		messageID1 := s.messageCreatedFromFile("user", mboxID1, "testdata/afternoon-meeting.eml")
		messageID2 := s.messageCreatedFromFile("user", mboxID1, "testdata/afternoon-meeting.eml")

		// Add the messages to a second mailbox.
		s.messageAdded("user", messageID1, mboxID2)
		s.messageAdded("user", messageID2, mboxID2)

		// Select in the mailbox.
		c[1].C("tag select mbox1").OK("tag")
		c[2].C("tag select mbox1").OK("tag")

		// First client expunges the messages.
		c[1].C(`tag store 1:* +flags (\deleted)`).OK("tag")
		c[1].C(`tag expunge`).OK("tag")

		// Second client has not been notified of the expunge.
		// The messages are in the deletion pool.
		c[2].C(`tag fetch 1:* (uid)`)
		c[2].Se(`* 1 FETCH (UID 1)`, `* 2 FETCH (UID 2)`)
		c[2].OK("tag")

		// Put the messages back in the mailbox.
		// They'll get new UIDs.
		s.messageAdded("user", messageID1, mboxID1)
		s.messageAdded("user", messageID2, mboxID1)

		// Flush the updates.
		s.flush("user")

		// Receive updates.
		c[2].C(`tag noop`).OK(`tag`)

		// Second client sees the messages have new UIDs.
		c[2].C(`tag fetch 1:* (uid)`)
		c[2].Se(`* 1 FETCH (UID 3)`, `* 2 FETCH (UID 4)`)
		c[2].OK("tag")
	})
}

func TestRemoteConcurrency(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.LabelID) {
		for mbox := 1; mbox <= 100; mbox++ {
			c.Cf(`tag create %v`, mbox).OK(`tag`)
		}

		s.flush("user")

		for mbox := 1; mbox <= 100; mbox++ {
			c.Cf(`tag move 1:* %v`, mbox).OK(`tag`)
			c.Cf(`tag select %v`, mbox).OK(`tag`)
		}

		s.flush("user")

		for mbox := 1; mbox <= 100; mbox++ {
			if mbox < 100 {
				c.Cf(`tag status %v (messages)`, mbox).Sxe(`MESSAGES 0`).OK(`tag`)
			} else {
				c.Cf(`tag status %v (messages)`, mbox).Sxe(`MESSAGES 100`).OK(`tag`)
			}
		}
	})
}
