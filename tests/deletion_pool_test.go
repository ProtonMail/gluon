package tests

import (
	"testing"
)

func TestDeletionPool(t *testing.T) {
	runManyToOneTestWithAuth(t, defaultServerOptions(t), []int{1, 2, 3, 4}, func(c map[int]*testConnection, s *testSession) {
		c[1].doAppend(`INBOX`, `To: 1@pm.me`).expect("OK")
		c[1].doAppend(`INBOX`, `To: 2@pm.me`).expect("OK")
		c[1].doAppend(`INBOX`, `To: 3@pm.me`).expect("OK")

		for _, i := range []int{1, 2, 4} {
			c[i].C("A006 SELECT INBOX")
			c[i].Se(`* 3 EXISTS`)
			c[i].Sxe("A006 OK")
		}

		// From session 1, we flag the 2nd message as deleted and expunge. Message is not listed anymore.
		c[1].C(`B001 STORE 2 +FLAGS (\Deleted)`)
		c[1].S(`* 2 FETCH (FLAGS (\Deleted \Recent))`) // update due to session 1
		c[1].Sx(`B001 OK`)
		c[1].C(`B002 EXPUNGE`)
		c[1].S(`* 2 EXPUNGE`)
		c[1].Sx(`B002 OK`)
		c[1].C(`B003 FETCH 1:* (UID)`)
		c[1].S(`* 1 FETCH (UID 1)`, `* 2 FETCH (UID 3)`)
		c[1].Sx(`B003 OK`)

		// From session 2, we first FETCH, the second message should still be there, flagged as deleted
		// Then we send a no-op, get a notification for the EXPUNGE and the message  will have been removed
		c[2].C(`C001 FETCH 1:* (FLAGS UID)`)
		c[2].S(`* 2 FETCH (FLAGS (\Deleted))`) // Queued snapshot update
		c[2].S(`* 1 FETCH (FLAGS () UID 1)`, `* 2 FETCH (FLAGS (\Deleted) UID 2)`, `* 3 FETCH (FLAGS () UID 3)`)
		c[2].Sx(`C001 OK \[EXPUNGEISSUED\][ \..]*`)
		c[2].C(`C002 NOOP`)
		c[2].S(`* 2 EXPUNGE`)
		c[2].Sx(`C002 OK`)
		c[2].C(`C003 FETCH 1:* (FLAGS UID)`)
		c[2].S(`* 1 FETCH (FLAGS () UID 1)`, `* 2 FETCH (FLAGS () UID 3)`)
		c[2].OK(`C003`)

		// We create a new snapshot of the INBOX, it only lists two messages
		c[3].C(`D001 SELECT INBOX`)
		c[3].Se(`* 2 EXISTS`)
		c[3].Sxe(`D001 OK`)

		// From session 4, we check that we still have 3 messages. Then we close the mailbox,
		// get the FETCH and EXPUNGE notification, and reselect the mailbox. The deleted message should not be there anymore
		c[4].C(`E001 FETCH 1:* (FLAGS UID)`)
		c[4].S(`* 2 FETCH (FLAGS (\Deleted))`, `* 1 FETCH (FLAGS () UID 1)`, `* 2 FETCH (FLAGS (\Deleted) UID 2)`, `* 3 FETCH (FLAGS () UID 3)`)
		c[4].Sxe(`E001 OK \[EXPUNGEISSUED\]`)
		c[4].C(`E002 CLOSE`).OK(`E002`)
		c[4].C(`E003 SELECT INBOX`)
		c[4].Sxe(`\* 2 EXISTS`)
		c[4].Sxe(`E003 OK`)
		c[4].C(`E004 FETCH 1:* (FLAGS UID)`)
		c[4].S(`* 1 FETCH (FLAGS () UID 1)`, `* 2 FETCH (FLAGS () UID 3)`)
		c[4].OK(`E004`)
	})
}

func TestExpungeIssued(t *testing.T) {
	runManyToOneTestWithAuth(t, defaultServerOptions(t), []int{1, 2}, func(c map[int]*testConnection, s *testSession) {
		c[1].doAppend(`INBOX`, `To: 1@pm.me`).expect("OK")
		c[1].doAppend(`INBOX`, `To: 2@pm.me`).expect("OK")
		c[1].doAppend(`INBOX`, `To: 3@pm.me`).expect("OK")

		// 2 snapshots of INBOX
		for i := 1; i <= 2; i++ {
			c[i].C("A001 SELECT INBOX")
			c[i].Se(`* 3 EXISTS`)
			c[i].Sxe("A001 OK")
		}

		// delete and expunge in snapshot 1
		c[1].C(`B001 STORE 2 +FLAGS (\Deleted)`)
		c[1].S(`* 2 FETCH (FLAGS (\Deleted \Recent))`) // update due to session 1
		c[1].Sx(`B001 OK`)
		c[1].C(`B002 EXPUNGE`)
		c[1].S(`* 2 EXPUNGE`)
		c[1].Sx(`B002 OK`)
		c[1].C(`B003 FETCH 1:* (UID)`)
		c[1].S(`* 1 FETCH (UID 1)`, `* 2 FETCH (UID 3)`)
		c[1].Sx(`B003 OK`)

		// FETCH, STORE and SEARCH from unnotified snapshot 2. we should get EXPUNGEISSUED even for untouched messages
		c[2].C(`C001 FETCH 1:* (FLAGS UID)`)
		c[2].S(`* 2 FETCH (FLAGS (\Deleted))`)
		c[2].S(`* 1 FETCH (FLAGS () UID 1)`, `* 2 FETCH (FLAGS (\Deleted) UID 2)`, `* 3 FETCH (FLAGS () UID 3)`)
		c[2].Sx(`C001 OK \[EXPUNGEISSUED\] command completed in`)

		c[2].C(`C002 FETCH 1 (FLAGS UID)`)
		c[2].S(`* 1 FETCH (FLAGS () UID 1)`)
		c[2].Sx(`C002 OK \[EXPUNGEISSUED\] command completed in`)

		c[2].C(`C003 STORE 2 +FLAGS (flag)`)
		c[2].S(`* 2 FETCH (FLAGS (\Deleted flag))`)
		c[2].Sx(`C003 OK \[EXPUNGEISSUED\] command completed in`)

		c[2].C(`C004 SEARCH DELETED`)
		c[2].S(`* SEARCH 2`)
		c[2].Sx(`C004 OK \[EXPUNGEISSUED\] command completed in`)

		c[2].C(`C005 NOOP`)
		c[2].S(`* 2 EXPUNGE`)
		c[2].Sx(`C005 OK`)
	})
}

func TestExpungeUpdate(t *testing.T) {
	runManyToOneTestWithAuth(t, defaultServerOptions(t), []int{1, 2}, func(c map[int]*testConnection, s *testSession) {
		c[1].doAppend(`INBOX`, `To: 1@pm.me`).expect("OK")
		c[1].doAppend(`INBOX`, `To: 2@pm.me`).expect("OK")

		// 2 snapshots of INBOX
		for i := 1; i <= 2; i++ {
			c[i].C(`A001 SELECT INBOX`)
			c[i].Se(`* 2 EXISTS`)
			c[i].Sxe(`A001 OK`)
		}

		// session 1 delete message 1 (STORE+EXPUNGE) then flag message 2 (now 1) as sent
		c[1].C(`A002 STORE 1 +FLAGS (\Deleted)`)
		c[1].S(`* 1 FETCH (FLAGS (\Deleted \Recent))`)
		c[1].Sx("A002 OK command completed in")
		c[1].C(`A003 EXPUNGE`)
		c[1].S(`* 1 EXPUNGE`)
		c[1].Sx(`A003 OK`)
		c[1].C(`A004 STORE 1 +FLAGS (\Seen)`)
		c[1].S(`* 1 FETCH (FLAGS (\Recent \Seen))`)
		c[1].Sx(`A004 OK command completed in`)

		// session 2 get notifications for the flags, but not the EXPUNGE.
		c[2].C(`B001 FETCH 1:* (UID FLAGS)`)
		c[2].S(
			`* 1 FETCH (FLAGS (\Deleted))`,
			`* 2 FETCH (FLAGS (\Seen))`,
		)
		c[2].S(
			`* 1 FETCH (UID 1 FLAGS (\Deleted))`,
			`* 2 FETCH (UID 2 FLAGS (\Seen))`,
		) // fetch results + updated from session 1
		c[2].Sx(`B001 OK \[EXPUNGEISSUED\] command completed in`)

		// session 2 can still change flags on the deleted message (which still resides in the deletion pool
		c[2].C(`B002 STORE 1 +FLAGS (\Flagged)`)
		c[2].S(`* 1 FETCH (FLAGS (\Deleted \Flagged))`)
		c[2].Sx(`B002 OK \[EXPUNGEISSUED\] command completed in`)

		// session 1 adds a message to the box and flags it as answered
		c[1].doAppend(`INBOX`, `To: 3@pm.me`).expect(`OK .*`)
		c[1].C(`A005 STORE 2 +FLAGS (\Answered)`)
		c[1].S(`* 2 FETCH (FLAGS (\Answered \Recent))`)
		c[1].Sx("A005 OK command completed in")

		// session 2 sees the new message
		c[2].C(`B003 FETCH 1:* (UID FLAGS)`)
		c[2].S(`* 3 EXISTS`)
		c[2].S(
			`* 1 FETCH (UID 1 FLAGS (\Deleted \Flagged))`,
			`* 2 FETCH (UID 2 FLAGS (\Seen))`,
			`* 3 FETCH (UID 3 FLAGS (\Answered))`,
		)
		c[2].Sx(`B003 OK \[EXPUNGEISSUED\] command completed in`)

		// session 2 performs a NOOP to get notified of the EXPUNGE from session 1
		c[2].C(`B004 NOOP`)
		c[2].S(`* 1 EXPUNGE`)
		c[2].Sx(`B004 OK`)
		c[2].C(`B005 FETCH 1:* (UID FLAGS)`)
		c[2].S(`* 1 FETCH (UID 2 FLAGS (\Seen))`, `* 2 FETCH (UID 3 FLAGS (\Answered))`)
		c[2].Sx(`B005 OK command completed in`)

		// close sessions
		for i := 1; i <= 2; i++ {
			c[i].C(`C001 CLOSE`)
			c[i].Sx(`C001 OK`)
		}
	})
}

func TestStatusOnUnnotifiedSnapshot(t *testing.T) {
	runManyToOneTestWithAuth(t, defaultServerOptions(t), []int{1, 2, 3}, func(c map[int]*testConnection, s *testSession) {
		// INBOX with 4 messages
		c[1].doAppend(`INBOX`, `To: 1@pm.me`, `\Seen`).expect("OK")
		c[1].doAppend(`INBOX`, `To: 2@pm.me`, `\Seen`).expect("OK")
		c[1].doAppend(`INBOX`, `To: 3@pm.me`).expect("OK")
		c[1].doAppend(`INBOX`, `To: 4@pm.me`).expect("OK")

		// 2 sessions with INBOX selected (-> 2 snapshots), session 3 is in Authenticated state (no mailbox selected).
		for i := 1; i <= 2; i++ {
			c[i].C(`A001 SELECT INBOX`)
			c[i].Se(`* 4 EXISTS`)
			c[i].Sxe(`A001 OK`)
		}

		// snapshot 1 deletes message 1 (STORE + EXPUNGE).
		c[1].C(`A002 STORE 1 +FLAGS (\Deleted)`)
		c[1].S(`* 1 FETCH (FLAGS (\Deleted \Recent \Seen))`)
		c[1].Sx(`A002 OK command completed in`)
		c[1].C(`A003 EXPUNGE`)
		c[1].S(`* 1 EXPUNGE`)
		c[1].S(`A003 OK EXPUNGE`)

		// session 3 calls status.
		c[3].C(`C001 STATUS INBOX (MESSAGES RECENT UNSEEN)`)
		c[3].S(`* STATUS "INBOX" (MESSAGES 3 RECENT 0 UNSEEN 2)`)
		c[3].S(`C001 OK STATUS`)

		// session 2 (which has INBOX selected) calls status and, it gets the updates and status for the messages.
		c[2].C(`B001 STATUS INBOX (MESSAGES RECENT UNSEEN)`)
		c[2].S(`* 1 FETCH (FLAGS (\Deleted \Seen))`)
		c[2].S(`* 1 EXPUNGE`)
		c[2].S(`* STATUS "INBOX" (MESSAGES 3 RECENT 0 UNSEEN 2)`)
		c[2].S(`B001 OK STATUS`)
	})
}

func TestDeletionFlagPropagation(t *testing.T) {
	runManyToOneTestWithAuth(t, defaultServerOptions(t), []int{1, 2}, func(c map[int]*testConnection, s *testSession) {
		origin, done := c[1].doCreateTempDir()
		defer done()

		// Create a message.
		c[1].doAppend(origin, `To: 1@pm.me`).expect(`OK`)

		destination, done := c[1].doCreateTempDir()
		defer done()

		// Client 1 is in origin, client 2 is in destination.
		c[1].Cf("A006 SELECT %v", origin).OK("A006")
		c[2].Cf("A006 SELECT %v", destination).OK("A006")

		// Copy the message to destination; there will be one message there.
		c[1].Cf(`A007 COPY 1 %v`, destination).OK("A007")
		c[2].Cf(`A006 STATUS %v (MESSAGES)`, destination).Sxe("MESSAGES 1").OK("A006")

		// Mark the message we copied as deleted. Only client 1 sees the flag.
		c[1].C(`A001 STORE 1 +FLAGS (\Deleted)`).OK("A001")
		c[1].C(`A005 FETCH 1 (FLAGS)`).Sxe(`FLAGS \(\\Deleted \\Recent\)`).OK("A005")
		c[2].C(`B005 FETCH 1 (FLAGS)`).Sxe(`FLAGS \(\\Recent\)`).OK("B005")

		// Expunge the message from origin; the message is now only in destination.
		c[1].C(`B002 EXPUNGE`).OK("B002")
		c[1].Cf(`A006 STATUS %v (MESSAGES)`, origin).Sxe("MESSAGES 0").OK("A006")
		c[2].Cf(`B006 STATUS %v (MESSAGES)`, destination).Sxe("MESSAGES 1").OK("B006")

		// The message in destination still doesn't have the deleted flag.
		c[2].C(`B005 FETCH 1 (FLAGS)`).Sxe(`FLAGS \(\\Recent\)`).OK("B005")
	})
}

func TestDeletionFlagPropagationIDLE(t *testing.T) {
	runManyToOneTestWithAuth(t, defaultServerOptions(t), []int{1, 2}, func(c map[int]*testConnection, s *testSession) {
		// Create a message.
		c[1].doAppend(`INBOX`, `To: 1@pm.me`).expect(`OK`)

		// Create a destination mailbox.
		c[1].C("A002 CREATE destination").OK("A002")

		// Client 1 is in inbox, client 2 is in destination.
		c[1].C("A006 SELECT INBOX").OK("A006")
		c[2].C("A006 SELECT destination").OK("A006")

		// Copy the message to destination; there will be one message there.
		c[1].C(`A007 COPY 1 destination`).OK("A007")
		c[2].C(`A006 STATUS destination (MESSAGES)`).Sxe("MESSAGES 1").OK("A006")

		// Begin IDLE in destination.
		c[2].C("A007 IDLE")
		c[2].S("+ Ready")

		// Delete the message from inbox (set \Deleted and expunge)
		c[1].C(`A001 STORE 1 +FLAGS (\Deleted)`)
		c[1].S(`* 1 FETCH (FLAGS (\Deleted \Recent))`)
		c[1].Sx(`A001 OK`)
		c[1].C(`B002 EXPUNGE`)
		c[1].S(`* 1 EXPUNGE`)
		c[1].Sx(`B002 OK`)

		// The client doing IDLE in destination shouldn't receive any updates as from its perspective nothing changed.
		c[2].C(`DONE`)
		c[2].Sx(`A007 OK`)
	})
}

func TestDeletionFlagPropagationMulti(t *testing.T) {
	runManyToOneTestWithAuth(t, defaultServerOptions(t), []int{1, 2, 3}, func(c map[int]*testConnection, s *testSession) {
		// Create a message.
		c[1].doAppend(`INBOX`, `To: 1@pm.me`).expect(`OK`)

		// Create two destination mailboxes.
		c[1].C("A002 CREATE mbox1").OK("A002")
		c[1].C("A002 CREATE mbox2").OK("A002")

		// Copy the message into both destination mailboxes.
		c[1].C("A006 SELECT INBOX").OK("A006")
		c[1].C("A006 COPY 1 mbox1").OK("A006")
		c[1].C("A006 COPY 1 mbox2").OK("A006")

		// Clients 1 and 2 select in mbox1/mbox2. Client 3 stays in inbox.
		c[1].C("A006 SELECT mbox1").OK("A006")
		c[2].C("A006 SELECT mbox2").OK("A006")
		c[3].C("A006 SELECT inbox").OK("A006")

		// Mark the message in mbox1 as deleted.
		c[1].C(`A001 STORE 1 +FLAGS (\Deleted)`).OK(`A001`)

		// Expect to receive no updates in mbox2/inbox.
		c[2].C(`A007 NOOP`)
		c[2].Sx(`A007 OK`)
		c[3].C(`A007 NOOP`)
		c[3].Sx(`A007 OK`)

		// Mark the message in mbox1 as custom1.
		c[1].C(`A001 STORE 1 +FLAGS (custom1)`).OK(`A001`)

		// Expect to receive the custom1 flag but not the deleted flag.
		// The recent flag should only be returned for mbox2.
		c[2].C(`A007 NOOP`)
		c[2].S(`* 1 FETCH (FLAGS (\Recent custom1))`)
		c[2].Sx(`A007 OK`)
		c[3].C(`A007 NOOP`)
		c[3].S(`* 1 FETCH (FLAGS (custom1))`)
		c[3].Sx(`A007 OK`)

		// Unmark the message in mbox1 as deleted.
		c[1].C(`A001 STORE 1 -FLAGS (\Deleted)`).OK(`A001`)

		// Expect to receive no updates in mbox2/inbox.
		c[2].C(`A007 NOOP`)
		c[2].Sx(`A007 OK`)
		c[3].C(`A007 NOOP`)
		c[3].Sx(`A007 OK`)
	})
}

func TestNoopReceivesPendingDeletionUpdates(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		// Create a mailbox.
		mailboxID := s.mailboxCreated("user", []string{"mbox"})

		// Create some messages in the mailbox.
		messageID1 := s.messageCreatedFromFile("user", mailboxID, "testdata/multipart-mixed.eml")
		messageID2 := s.messageCreatedFromFile("user", mailboxID, "testdata/afternoon-meeting.eml")

		// Create a snapshot by selecting in the mailbox.
		c.C(`A001 select mbox`).OK(`A001`)

		// Remove the messages externally; they're now in the deletion pool.
		s.messageRemoved("user", messageID1, mailboxID)
		s.messageRemoved("user", messageID2, mailboxID)

		// Noop should process their deletion.
		c.C(`A002 noop`)
		c.S(`* 1 EXPUNGE`, `* 1 EXPUNGE`)
		c.Sx(`A002 OK`)
	})
}

func TestMessageErasedFromDB(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		// Create a mailbox.
		mailboxID := s.mailboxCreated("user", []string{"mbox"})

		// Create some messages in the mailbox.
		messageID1 := s.messageCreatedFromFile("user", mailboxID, "testdata/multipart-mixed.eml")
		messageID2 := s.messageCreatedFromFile("user", mailboxID, "testdata/afternoon-meeting.eml")

		// Create a snapshot by selecting in the mailbox.
		c.C(`A001 select mbox`).OK(`A001`)

		dbCheckUserMessageCount(s, "user", 2)

		// Messages marked for deletion externally
		s.messageDeleted("user", messageID1)
		s.messageDeleted("user", messageID2)

		waiter := newEventWaiter(s.server)
		defer waiter.close()

		// Noop should process their deletion.
		c.C(`A002 LOGOUT`)
		c.S(`* BYE`)
		c.Sx(`A002 OK`)

		waiter.waitEndOfSession()
		dbCheckUserMessageCount(s, "user", 0)
	})
}

// This test checks whether any messages from a previous server that are written to the db and are cleared
// on the next startup. We force the server to use the same directories and state to check for this.
func TestMessageErasedFromDBOnStartup(t *testing.T) {
	options := defaultServerOptions(t, withDataDir(t.TempDir()))

	runOneToOneTest(t, options, func(c *testConnection, s *testSession) {
		// Create a mailbox.
		mailboxID := s.mailboxCreated("user", []string{"mbox"})

		// Create some messages in the mailbox.
		messageID1 := s.messageCreatedFromFile("user", mailboxID, "testdata/multipart-mixed.eml")

		waiter := newEventWaiter(s.server)
		defer waiter.close()

		// Noop should process their deletion.
		c.C(`A002 LOGOUT`)
		c.S(`* BYE`)
		c.Sx(`A002 OK LOGOUT`)
		waiter.waitEndOfSession()
		dbCheckUserMessageCount(s, "user", 1)

		// delete message
		s.messageDeleted("user", messageID1)
		dbCheckUserMessageCount(s, "user", 1)
	})

	runOneToOneTest(t, options, func(c *testConnection, s *testSession) {
		// Message should have been removed now.
		dbCheckUserMessageCount(s, "user", 0)
	})
}

func TestMessageErasedFromDBWithMany(t *testing.T) {
	runManyToOneTestWithAuth(t, defaultServerOptions(t), []int{1, 2}, func(c map[int]*testConnection, s *testSession) {
		// Create a mailbox.
		mailboxID := s.mailboxCreated("user", []string{"mbox"})

		// Create some messages in the mailbox.
		messageID1 := s.messageCreatedFromFile("user", mailboxID, "testdata/multipart-mixed.eml")
		s.messageCreatedFromFile("user", mailboxID, "testdata/afternoon-meeting.eml")

		// Create a snapshot by selecting in the mailbox.
		c[1].C("A002 SELECT mbox").OK("A002")
		c[2].C("A002 SELECT mbox").OK("A002")

		dbCheckUserMessageCount(s, "user", 2)

		// Message marked for deletion externally
		s.messageDeleted("user", messageID1)

		waiter := newEventWaiter(s.server)
		defer waiter.close()

		// Logout client 1.
		c[1].C(`A002 LOGOUT`)
		c[1].S(`* BYE`)
		c[1].Sx(`A002 OK LOGOUT`)

		// Ensure session is properly finished its exit work to ensure the database writes take place.
		waiter.waitEndOfSession()

		// Message should still be in the db as the other client still has an active state instance.
		dbCheckUserMessageCount(s, "user", 2)

		c[2].C(`A002 LOGOUT`)
		// c[2].S(`* 2 EXISTS`)
		// c[2].S(`* 1 RECENT`)
		c[2].S(`* BYE`)
		c[2].Sx(`A002 OK LOGOUT`)

		// Ensure session is properly finished its exit work to ensure the database writes take place.
		waiter.waitEndOfSession()

		// The message should now be deleted.
		dbCheckUserMessageCount(s, "user", 1)
	})
}
