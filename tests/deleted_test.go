package tests

import (
	"testing"
)

func TestDeleted(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		// Create two mailboxes.
		c.C("b001 CREATE mbox1")
		c.S("b001 OK CREATE")
		c.C("b001 CREATE mbox2")
		c.S("b001 OK CREATE")

		// Create a message in mbox1.
		c.doAppend(`mbox1`, `To: 1@pm.me`, `\Seen`).expect("OK")
		c.doAppend(`mbox1`, `To: 2@pm.me`, `\Seen`).expect("OK")
		c.C(`A002 SELECT mbox1`)
		c.Se(`A002 OK [READ-WRITE] SELECT`)

		// Copy messages 1 to mbox2 and flag it as deleted in mbox 1.
		c.C(`A003 COPY 1 mbox2`)
		c.Sx(`A003 OK .*`)
		c.C(`A004 STORE 1 +FLAGS (\Deleted)`)
		c.S(`* 1 FETCH (FLAGS (\Deleted \Recent \Seen))`)
		c.OK("A004")
		c.C(`B001 FETCH 1 (FLAGS)`)
		c.S(`* 1 FETCH (FLAGS (\Deleted \Recent \Seen))`)
		c.OK("B001")
		c.C(`B002 FETCH 2 (FLAGS)`)
		c.S(`* 2 FETCH (FLAGS (\Recent \Seen))`)
		c.OK("B002")

		// Check that the copy in mbox2 does not have the flag \Deleted.
		c.C(`A005 SELECT mbox2`)
		c.Se(`* 1 EXISTS`)
		c.Se(`A005 OK [READ-WRITE] SELECT`)
		c.C(`A006 FETCH 1 (FLAGS)`)
		c.S(`* 1 FETCH (FLAGS (\Recent \Seen))`)
		c.OK(`A006`)

		// Expunge the copy in mbox1.
		// The message no longer has the recent flag.
		c.C(`A007 SELECT mbox1`)
		c.Se(`* 2 EXISTS`)
		c.Se(`A007 OK [READ-WRITE] SELECT`)
		c.C(`A008 EXPUNGE`)
		c.S(`* 1 EXPUNGE`)
		c.OK(`A008`)
		c.C(`A009 STATUS mbox1 (MESSAGES)`)
		c.S(`* STATUS "mbox1" (MESSAGES 1)`)
		c.OK(`A009`)

		// Check that the message is still in mbox2
		// The message no longer has the recent flag.
		c.C(`A00A SELECT mbox2`)
		c.Se(`* 1 EXISTS`)
		c.Se(`A00A OK [READ-WRITE] SELECT`)

		// Flag, unflag, expunge and check the message is still there.
		c.C(`A00B STORE 1 +FLAGS (\Deleted)`)
		c.S(`* 1 FETCH (FLAGS (\Deleted \Seen))`)
		c.OK(`A00B`)
		c.C(`A00C STORE 1 -FLAGS (\Deleted)`)
		c.S(`* 1 FETCH (FLAGS (\Seen))`)
		c.OK(`A00C`)
		c.C(`A00D EXPUNGE`)
		c.S(`A00D OK EXPUNGE`)
		c.C(`A00E STATUS mbox1 (MESSAGES)`)
		c.S(`* STATUS "mbox1" (MESSAGES 1)`)
		c.S(`A00E OK STATUS`)
	})
}

func TestUIDDeleted(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		// Create two mailboxes
		c.C("b001 CREATE mbox1")
		c.S("b001 OK CREATE")
		c.C("b001 CREATE mbox2")
		c.S("b001 OK CREATE")

		// Create a message in mbox1
		c.doAppend(`mbox1`, `To: 1@pm.me`, `\Seen`).expect("OK")
		c.doAppend(`mbox1`, `To: 2@pm.me`, `\Seen`).expect("OK")
		c.C(`A002 SELECT mbox1`)
		c.Se(`A002 OK [READ-WRITE] SELECT`)

		// Copy message 2 to mbox2 and flag it as deleted in mbox 1
		c.C(`A003 UID COPY 2 mbox2`)
		c.Sx(`A003 OK .*`)
		c.C(`A004 UID STORE 2 +FLAGS (\Deleted)`)
		c.S(`* 2 FETCH (FLAGS (\Deleted \Recent \Seen) UID 2)`)
		c.OK(`A004`)

		// Check that the copy in mbox2 is does not have the flag \Deleted
		c.C(`A005 SELECT mbox2`)
		c.Se(`* 1 EXISTS`)
		c.Se(`A005 OK [READ-WRITE] SELECT`)
		c.C(`A006 UID FETCH 1 (FLAGS)`)
		c.S(`* 1 FETCH (FLAGS (\Recent \Seen) UID 1)`)
		c.OK(`A006`)

		// Expunge the copy in mbox1
		c.C(`A007 SELECT mbox1`)
		c.Se(`* 2 EXISTS`)
		c.Se(`A007 OK [READ-WRITE] SELECT`)
		c.C(`A008 EXPUNGE`)
		c.S(`* 2 EXPUNGE`)
		c.Sx(`A008 OK .*`)
		c.C(`A009 STATUS mbox1 (MESSAGES)`)
		c.S(`* STATUS "mbox1" (MESSAGES 1)`)
		c.S(`A009 OK STATUS`)

		// Check that the message is still in mbox2
		c.C(`A00A SELECT mbox2`)
		c.Se(`* 1 EXISTS`)
		c.Se(`A00A OK [READ-WRITE] SELECT`)

		// Flag,unflag, expunge and check the message is still there.
		c.C(`A00B UID STORE 1 +FLAGS (\Deleted)`)
		c.S(`* 1 FETCH (FLAGS (\Deleted \Seen) UID 1)`)
		c.OK(`A00B`)
		c.C(`A00C UID STORE 1 -FLAGS (\Deleted)`)
		c.S(`* 1 FETCH (FLAGS (\Seen) UID 1)`)
		c.OK(`A00C`)
		c.C(`A00D EXPUNGE`)
		c.S(`A00D OK EXPUNGE`)
		c.C(`A00E STATUS mbox1 (MESSAGES)`)
		c.S(`* STATUS "mbox1" (MESSAGES 1)`)
		c.S(`A00E OK STATUS`)
	})
}
