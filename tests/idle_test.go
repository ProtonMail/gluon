package tests

import (
	"testing"
)

func TestIDLEExistsUpdates(t *testing.T) {
	runManyToOneTestWithAuth(t, defaultServerOptions(t), []int{1, 2}, func(c map[int]*testConnection, s *testSession) {
		// First client selects in INBOX to receive EXISTS update.
		c[1].C("A006 select INBOX")
		c[1].Se("A006 OK [READ-WRITE] (^_^)")

		// First client starts to IDLE.
		c[1].C("A007 IDLE")
		c[1].S("+ (*_*)")

		// Second client appends to INBOX to generate EXISTS updates.
		// The client is not selected and thus doesn't itself receive responses.
		c[2].doAppend(`INBOX`, `To: 1@pm.me`, `\Seen`).expect("OK")
		c[2].doAppend(`INBOX`, `To: 2@pm.me`, `\Seen`).expect("OK")

		// First client receives the EXISTS and RECENT updates while idling.
		c[1].S(`* 1 EXISTS`, `* 1 RECENT`, `* 2 EXISTS`, `* 2 RECENT`)

		// First client stops idling.
		c[1].C("DONE")
		c[1].S(`A007 OK (^_^)`)

		// Further stuff doesn't trigger any issues.
		c[2].doAppend(`INBOX`, `To: 3@pm.me`, `\Seen`).expect("OK")
	})
}

func TestIDLEPendingUpdates(t *testing.T) {
	runManyToOneTestWithData(t, defaultServerOptions(t), []int{1, 2}, func(c map[int]*testConnection, s *testSession, _, _ string) {
		c[1].C("A001 select INBOX").OK("A001")

		// Generate some pending updates.
		c[2].C("B001 UID MOVE 1,2,3 INBOX").OK("B001")

		// Begin IDLE.
		c[1].C("A002 IDLE").S("+ (*_*)")

		// Generate some additional updates.
		c[2].C("B002 UID MOVE 4,5,6 INBOX").OK("B002")

		// Pending updates are first flushed.
		c[1].Se(`* 1 EXISTS`, `* 2 EXISTS`, `* 3 EXISTS`)

		// IDLE updates are first second.
		c[1].Se(`* 4 EXISTS`, `* 5 EXISTS`, `* 6 EXISTS`)

		// Stop IDLE.
		c[1].C("DONE").OK("A002")
	})
}
