package tests

import (
	"testing"
)

func TestRecentSelect(t *testing.T) {
	runManyToOneTestWithAuth(t, defaultServerOptions(t), []int{1, 2, 3}, func(c map[int]*testConnection, _ *testSession) {
		// Client 1 appends a new message to INBOX.
		c[1].doAppend(`INBOX`, `To: 1@pm.me`).expect("OK")

		// Client 2 is the first to be notified of the new message; it appears as recent to client 2.
		c[2].C("A006 select INBOX")
		c[2].Se(`* 1 EXISTS`, `* 1 RECENT`).OK("A006")

		// Client 3 is the second to be notified of the new message; it does not appear as recent to client 3.
		c[3].C("A006 select INBOX")
		c[3].Se(`* 1 EXISTS`, `* 0 RECENT`).OK("A006")
	})
}

func TestRecentStatus(t *testing.T) {
	runManyToOneTestWithAuth(t, defaultServerOptions(t), []int{1, 2, 3}, func(c map[int]*testConnection, _ *testSession) {
		// Client 1 appends a new message to INBOX.
		c[1].doAppend("INBOX", `To: 1@pm.me`).expect("OK")

		// Client 2 is the first to be notified of the new message; it appears as recent to client 2.
		// Calling STATUS must not change the value of RECENT.
		c[2].C("A006 STATUS INBOX (RECENT)")
		c[2].S(`* STATUS "INBOX" (RECENT 1)`).OK("A006")

		// Client 3 is the second to be notified of the new message; it still appears as recent to client 3.
		c[3].C("A006 STATUS INBOX (RECENT)")
		c[3].S(`* STATUS "INBOX" (RECENT 1)`).OK("A006")
	})
}

func TestRecentFetch(t *testing.T) {
	runManyToOneTestWithAuth(t, defaultServerOptions(t), []int{1, 2, 3}, func(c map[int]*testConnection, _ *testSession) {
		// Client 1 appends a new message to INBOX.
		c[1].doAppend(`INBOX`, `To: 1@pm.me`).expect("OK")

		// Client 2 is the first to be notified of the new message; it appears as recent to client 2.
		c[2].C("A006 select INBOX")
		c[2].Se(`* 1 EXISTS`, `* 1 RECENT`).OK("A006")

		// Client 2 fetches the message; it is still recent to client 2.
		c[2].C("A006 fetch 1 (UID FLAGS)")
		c[2].Se(`* 1 FETCH (UID 1 FLAGS (\Recent))`).OK("A006")

		// Client 2 fetches the message again; it is still recent to client 2.
		c[2].C("A006 fetch 1 (UID FLAGS)")
		c[2].Se(`* 1 FETCH (UID 1 FLAGS (\Recent))`).OK("A006")

		// Client 3 is the second to be notified of the new message; it no longer appears as recent to client 3.
		c[3].C("A006 select INBOX")
		c[3].Se(`* 1 EXISTS`, `* 0 RECENT`).OK("A006")

		// Client 3 fetches the message; it is not recent to client 3.
		c[3].C("A006 fetch 1 (UID FLAGS)")
		c[3].Se(`* 1 FETCH (UID 1 FLAGS ())`).OK("A006")
	})
}

func TestRecentAppend(t *testing.T) {
	runManyToOneTestWithAuth(t, defaultServerOptions(t), []int{1, 2, 3}, func(c map[int]*testConnection, _ *testSession) {
		// Client 1 appends a new message to INBOX.
		c[1].doAppend(`INBOX`, `To: 1@pm.me`).expect("OK")

		// Client 2 is the first to be notified of the new message; it appears as recent to client 2.
		c[2].C("A006 select INBOX")
		c[2].Se(`* 1 EXISTS`, `* 1 RECENT`).OK("A006")

		// Client 2 then appends a second message to the mailbox while selected.
		// As it was the first client to perform the operation, it sees the message as recent.
		c[2].doAppend(`INBOX`, `To: 2@pm.me`).expect("OK")
		c[2].C("A006 fetch 2 (UID FLAGS)")
		c[2].Se(`* 2 FETCH (UID 2 FLAGS (\Recent))`).OK("A006")
	})
}

func TestRecentStore(t *testing.T) {
	runManyToOneTestWithAuth(t, defaultServerOptions(t), []int{1, 2}, func(c map[int]*testConnection, _ *testSession) {
		mbox, done := c[1].doCreateTempDir()
		defer done()

		// Create a message in mbox.
		c[1].doAppend(mbox, `To: 1@pm.me`).expect(`OK`)

		// Select in the mailbox.
		c[1].Cf(`A002 SELECT %v`, mbox).OK(`A002`)

		// Set the message as deleted.
		c[1].C(`A004 STORE 1 +FLAGS (\Deleted)`)
		c[1].S(`* 1 FETCH (FLAGS (\Deleted \Recent))`)
		c[1].OK(`A004`)

		// The message still has the recent flag when fetching.
		c[1].C("A006 FETCH 1 (UID FLAGS)")
		c[1].Se(`* 1 FETCH (UID 1 FLAGS (\Deleted \Recent))`)
		c[1].OK("A006")

		// Select in the mailbox.
		c[2].Cf(`A002 SELECT %v`, mbox).OK(`A002`)

		// The message does not have the recent flag when fetching because this client wasn't first notified.
		c[2].C("A006 FETCH 1 (UID FLAGS)")
		c[2].Se(`* 1 FETCH (UID 1 FLAGS (\Deleted))`)
		c[2].OK("A006")
	})
}

func TestRecentExists(t *testing.T) {
	runManyToOneTestWithAuth(t, defaultServerOptions(t), []int{1, 2, 3}, func(c map[int]*testConnection, _ *testSession) {
		// Client 1 and B both select INBOX.
		c[1].C("A006 select INBOX").OK("A006")
		c[2].C("A006 select INBOX").OK("A006")

		// Client 3 appends a new message to INBOX.
		c[3].doAppend(`INBOX`, `To: 1@pm.me`).expect("OK")

		// Client 1 is notified of the new message. No recent is sent as client 2 still has the mailbox selected.
		c[1].C("A007 NOOP")
		c[1].S("* 1 EXISTS").OK("A007")

		// Client 2 is notified of the new message second; it does not appear as recent to client 2.
		c[2].C("A007 NOOP")
		c[2].S("* 1 EXISTS").OK("A007")
	})
}

func TestRecentIDLEExists(t *testing.T) {
	runManyToOneTestWithAuth(t, defaultServerOptions(t), []int{1, 2}, func(c map[int]*testConnection, _ *testSession) {
		// Client 1 selects INBOX and IDLEs.
		c[1].C("A006 select INBOX").OK("A006")
		c[1].C("A007 IDLE")
		c[1].S("+ Ready")

		// Client 2 appends two new messages to INBOX.
		c[2].doAppend(`INBOX`, `To: 1@pm.me`).expect("OK")
		c[2].doAppend(`INBOX`, `To: 2@pm.me`).expect("OK")

		// Client 1 receives EXISTS and RECENT updates.
		c[1].S(`* 1 EXISTS`, `* 1 RECENT`, `* 2 EXISTS`, `* 2 RECENT`)
		c[1].C("DONE")
		c[1].OK(`A007`)
	})
}

func TestRecentIDLEExpunge(t *testing.T) {
	runManyToOneTestWithAuth(t, defaultServerOptions(t), []int{1, 2}, func(c map[int]*testConnection, _ *testSession) {
		// Client 1 creates a second mailbox and begins to IDLE inside it.
		c[1].C("A002 CREATE folder").OK("A002")
		c[1].C("A006 select folder").OK("A006")
		c[1].C("A007 IDLE")
		c[1].S("+ Ready")

		// Client 2 appends two new messages to INBOX.
		c[2].doAppend(`INBOX`, `To: 1@pm.me`).expect("OK")
		c[2].doAppend(`INBOX`, `To: 1@pm.me`).expect("OK")

		// Client 2 moves those two messages to the other folder.
		c[2].C("A006 select INBOX").OK("A006")
		c[2].C("A006 move 1:* folder").OK("A006")

		// Client 1 receives EXISTS. Since Client 2 still has the mailbox selected, recent updates are not sent.
		c[1].S(`* 1 EXISTS`, `* 2 EXISTS`)

		// Client 2 moves those two messages back to INBOX.
		// In doing so, it sees the messages in the folder; they are no longer recent.
		c[2].C("A006 select folder").OK("A006")
		c[2].C("A006 move 1:* INBOX").OK("A006")

		// Client 1 receives EXPUNGE updates for those messages.
		// It receives no RECENT updates because client 2 already saw them.
		// The order of expunge results cannot be guaranteed (MOVE is handled in random order).
		c[1].Sx(`\* \d EXPUNGE`, `\* \d EXPUNGE`)
		c[1].C("DONE")
		c[1].OK(`A007`)
	})
}
