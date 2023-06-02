package tests

import (
	"testing"
)

func TestCreateMulti(t *testing.T) {
	runManyToOneTestWithAuth(t, defaultServerOptions(t), []int{1, 2}, func(c map[int]*testConnection, _ *testSession) {
		c[1].C("A003 CREATE owatagusiam")
		c[1].OK("A003")

		c[2].C("A003 CREATE owatagusiam")
		c[2].S("A003 NO a mailbox with that name already exists")
	})
}

func TestExistsUpdates(t *testing.T) {
	runManyToOneTestWithAuth(t, defaultServerOptions(t), []int{1, 2}, func(c map[int]*testConnection, _ *testSession) {
		// First client selects in INBOX to receive EXISTS update.
		c[1].C("A006 select INBOX")
		c[1].Se("A006 OK [READ-WRITE] SELECT")

		// Second client appends to INBOX to generate EXISTS update.
		c[2].doAppend(`INBOX`, buildRFC5322TestLiteral(`To: 1@pm.me`), `\Seen`).expect("OK")

		// First client receives the EXISTS update. Apply RECENT update is also received.
		c[1].C("b001 noop")
		c[1].S(`* 1 EXISTS`)
		c[1].S(`* 1 RECENT`)
		c[1].OK("b001")
	})
}

func TestExistsUpdatesInSeparateMailboxes(t *testing.T) {
	runManyToOneTestWithAuth(t, defaultServerOptions(t), []int{1, 2}, func(c map[int]*testConnection, _ *testSession) {
		c[1].C("A003 CREATE owatagusiam")
		c[1].S("A003 OK CREATE")

		// First client selects in owatagusiam to ignore EXISTS updates from INBOX.
		c[1].C("A006 select owatagusiam")
		c[1].Se("A006 OK [READ-WRITE] SELECT")

		// Second client appends to INBOX to generate EXISTS update.
		c[2].doAppend(`INBOX`, buildRFC5322TestLiteral(`To: 1@pm.me`), `\Seen`).expect("OK")

		// First client does not receive the EXISTS update from INBOX.
		c[1].C("b001 noop")
		c[1].OK("b001")
	})
}

func TestFetchUpdates(t *testing.T) {
	runManyToOneTestWithAuth(t, defaultServerOptions(t), []int{1, 2}, func(c map[int]*testConnection, s *testSession) {
		c[1].doAppend(`INBOX`, buildRFC5322TestLiteral(`To: 1@pm.me`), `\Seen`).expect("OK")

		// First client selects in INBOX to receive FETCH update.
		c[1].C("A006 select INBOX")
		c[1].Se("A006 OK [READ-WRITE] SELECT")

		s.flush("user")

		// Second client selects in INBOX and then sets some flags to generate a FETCH update.
		c[2].C("b006 select INBOX")
		c[2].Se("b006 OK [READ-WRITE] SELECT")

		c[2].C(`B007 STORE 1 +FLAGS (\Deleted)`)
		c[2].S(`* 1 FETCH (FLAGS (\Deleted \Seen))`)
		c[2].Sx("B007 OK .*")

		s.flush("user")

		// First client receives the FETCH update.
		c[1].C("c001 noop")
		c[1].S(`* 1 FETCH (FLAGS (\Deleted \Recent \Seen))`)
		c[1].OK("c001")
	})
}

func TestExpungeUpdates(t *testing.T) {
	runManyToOneTestWithAuth(t, defaultServerOptions(t), []int{1, 2}, func(c map[int]*testConnection, _ *testSession) {
		// Generate three messages, the first two unseen, the third seen.
		c[1].doAppend(`INBOX`, buildRFC5322TestLiteral(`To: 1@pm.me`)).expect("OK")
		c[1].doAppend(`INBOX`, buildRFC5322TestLiteral(`To: 1@pm.me`)).expect("OK")
		c[1].doAppend(`INBOX`, buildRFC5322TestLiteral(`To: 1@pm.me`), `\Seen`).expect("OK")

		// Both clients select in inbox.
		c[1].C("A006 select INBOX")
		c[1].Se("A006 OK [READ-WRITE] SELECT")

		c[2].C("A007 select INBOX")
		c[2].Se("A007 OK [READ-WRITE] SELECT")

		// For both clients, the message with sequence number 3 is seen.
		c[1].C(`A005 FETCH 3 (FLAGS UID)`)
		c[1].S(`* 3 FETCH (FLAGS (\Recent \Seen) UID 3)`)
		c[1].OK(`A005`)
		c[2].C(`B005 FETCH 3 (FLAGS UID)`)
		c[2].S(`* 3 FETCH (FLAGS (\Seen) UID 3)`)
		c[2].OK(`B005`)

		// First client marks the first message as deleted.
		c[1].C(`B003 STORE 1 +FLAGS (\Deleted)`)
		c[1].S(`* 1 FETCH (FLAGS (\Deleted \Recent))`)
		c[1].OK("B003")

		// Second client sees the flag has been changed.
		c[2].C("c001 noop")
		c[2].S(`* 1 FETCH (FLAGS (\Deleted))`)
		c[2].OK("c001")

		// First client expunges the first message (seq numbers are shifted down by 1).
		c[1].C(`B202 EXPUNGE`)
		c[1].S(`* 1 EXPUNGE`)
		c[1].OK("B202")

		// Second client doesn't yet know that the messages were expunged
		// and it still thinks the seen message has seq 3 / uid 2
		// (actually, it was decremented, so it should now have seq 2 / uid 2)
		c[2].C(`B006 FETCH 3 (FLAGS UID)`)
		c[2].S(`* 3 FETCH (FLAGS (\Seen) UID 3)`)
		c[2].OK(`B006`)

		// Second client then does noop and gets the expunge update.
		// Its seqs are then decremented; the seen message should now have seq 2 / uid 2.
		c[2].C("c002 noop")
		c[2].S(`* 1 EXPUNGE`)
		c[2].OK("c002")
		c[2].C(`B007 FETCH 2 (FLAGS UID)`)
		c[2].S(`* 2 FETCH (FLAGS (\Seen) UID 3)`)
		c[2].OK(`B007`)
	})
}

func TestSequenceNumbersPerSession(t *testing.T) {
	runManyToOneTestWithAuth(t, defaultServerOptions(t), []int{1, 2}, func(c map[int]*testConnection, s *testSession) {
		// Generate five messages.
		c[1].doAppend(`inbox`, buildRFC5322TestLiteral(`To: 1@pm.me`)).expect("OK")
		c[1].doAppend(`inbox`, buildRFC5322TestLiteral(`To: 2@pm.me`)).expect("OK")
		c[1].doAppend(`inbox`, buildRFC5322TestLiteral(`To: 3@pm.me`)).expect("OK")
		c[1].doAppend(`inbox`, buildRFC5322TestLiteral(`To: 4@pm.me`)).expect("OK")
		c[1].doAppend(`inbox`, buildRFC5322TestLiteral(`To: 5@pm.me`)).expect("OK")

		// Both clients select in inbox.
		c[1].C("tag select inbox").OK("tag")
		c[2].C("tag select inbox").OK("tag")

		// Both clients initially see the same sequence numbers.
		c[1].C(`tag fetch 1:* (uid)`).Se(
			`* 1 FETCH (UID 1)`,
			`* 2 FETCH (UID 2)`,
			`* 3 FETCH (UID 3)`,
			`* 4 FETCH (UID 4)`,
			`* 5 FETCH (UID 5)`,
		).OK(`tag`)
		c[2].C(`tag fetch 1:* (uid)`).Se(
			`* 1 FETCH (UID 1)`,
			`* 2 FETCH (UID 2)`,
			`* 3 FETCH (UID 3)`,
			`* 4 FETCH (UID 4)`,
			`* 5 FETCH (UID 5)`,
		).OK(`tag`)

		// Expunge the first three messages with client 1.
		c[1].C(`tag store 1:3 +flags (\deleted)`).OK(`tag`)
		c[1].C(`tag expunge`).OK(`tag`)

		// Client 1 now only sees the last two messages; they now have sequence numbers 1 and 2.
		c[1].C(`tag fetch 1:* (uid)`).Se(
			`* 1 FETCH (UID 4)`,
			`* 2 FETCH (UID 5)`,
		).OK(`tag`)

		// However, client 2 doesn't know these messages have been deleted; it still sees all messages.
		c[2].C(`tag fetch 1:* (uid)`).Se(
			`* 1 FETCH (UID 1)`,
			`* 2 FETCH (UID 2)`,
			`* 3 FETCH (UID 3)`,
			`* 4 FETCH (UID 4)`,
			`* 5 FETCH (UID 5)`,
		).OK(`tag`)

		// Client 2 then becomes aware that these messages have been deleted.
		// (EXPUNGE can be performed in any order hence the regex here)
		c[2].C("tag noop").Sxe(
			`\* \d+ EXPUNGE`,
			`\* \d+ EXPUNGE`,
			`\* \d+ EXPUNGE`,
		).OK("tag")

		// Now that client 2 is aware of these messages having been expunged, it also only sees the last two messages.
		c[2].C(`tag fetch 1:* (uid)`).Se(
			`* 1 FETCH (UID 4)`,
			`* 2 FETCH (UID 5)`,
		).OK(`tag`)
	})
}

func TestAddFlagsToExpungedMessage(t *testing.T) {
	runManyToOneTestWithAuth(t, defaultServerOptions(t), []int{1, 2}, func(c map[int]*testConnection, s *testSession) {
		c[1].doAppend(`inbox`, buildRFC5322TestLiteral(`To: 1@pm.me`)).expect("OK")

		// Both clients select in inbox.
		c[1].C("tag select inbox").OK("tag")
		c[2].C("tag select inbox").OK("tag")

		// Both clients see one message in inbox.
		c[1].C(`tag status inbox (messages)`).Sxe(`MESSAGES 1`).OK(`tag`)
		c[2].C(`tag status inbox (messages)`).Sxe(`MESSAGES 1`).OK(`tag`)

		// Expunge the first message with client 1.
		c[1].C(`tag store 1 +flags (\deleted)`).OK(`tag`)
		c[1].C(`tag expunge`).OK(`tag`)
		c[1].C(`tag status inbox (messages)`).Sxe(`MESSAGES 0`).OK(`tag`)

		// Client 2 has not been notified of the expunge; setting flags must succeed.
		c[2].C(`tag store 1 +flags (\seen)`).OK(`tag`)

		// When client 2 performs noop, it finally receives the expunge response.
		c[2].C(`tag noop`).S(`* 1 EXPUNGE`).OK("tag")
		c[2].C(`tag status inbox (messages)`).Sxe(`MESSAGES 0`).OK(`tag`)
	})
}
