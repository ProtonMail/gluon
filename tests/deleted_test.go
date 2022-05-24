package tests

import (
	"testing"
)

func TestDeleted(t *testing.T) {
	runOneToOneTestWithAuth(t, "user", "pass", "/", func(c *testConnection, _ *testSession) {
		// Create two mailboxes.
		c.C("b001 CREATE mbox1")
		c.S("b001 OK (^_^)")
		c.C("b001 CREATE mbox2")
		c.S("b001 OK (^_^)")

		// Create a message in mbox1.
		c.doAppend(`mbox1`, `To: 1@pm.me`, `\Seen`).expect("OK")
		c.doAppend(`mbox1`, `To: 2@pm.me`, `\Seen`).expect("OK")
		c.C(`A002 SELECT mbox1`)
		c.Se(`A002 OK [READ-WRITE] (^_^)`)

		// Copy messages 1 to mbox2 and flag it as deleted in mbox 1.
		c.C(`A003 COPY 1 mbox2`)
		c.Sx(`A003 OK .*`)
		c.C(`A004 STORE 1 +FLAGS (\Deleted)`)
		c.S(`* 1 FETCH (FLAGS (\Deleted \Recent \Seen))`)
		c.Sx(`A004 OK .* command completed in .*`)
		c.C(`B001 FETCH 1 (FLAGS)`)
		c.S(`* 1 FETCH (FLAGS (\Deleted \Recent \Seen))`)
		c.Sx(`B001 OK .* command completed in`)
		c.C(`B002 FETCH 2 (FLAGS)`)
		c.S(`* 2 FETCH (FLAGS (\Recent \Seen))`)
		c.Sx(`B002 OK .* command completed in`)

		// Check that the copy in mbox2 does not have the flag \Deleted.
		c.C(`A005 SELECT mbox2`)
		c.Se(`* 1 EXISTS`)
		c.Se(`A005 OK [READ-WRITE] (^_^)`)
		c.C(`A006 FETCH 1 (FLAGS)`)
		c.S(`* 1 FETCH (FLAGS (\Recent \Seen))`)
		c.Sx(`A006 OK .* command completed in .*`)

		// Expunge the copy in mbox1.
		// The message no longer has the recent flag.
		c.C(`A007 SELECT mbox1`)
		c.Se(`* 2 EXISTS`)
		c.Se(`A007 OK [READ-WRITE] (^_^)`)
		c.C(`A008 EXPUNGE`)
		c.S(`* 1 EXPUNGE`)
		c.Sx(`A008 OK .*`)
		c.C(`A009 STATUS mbox1 (MESSAGES)`)
		c.S(`* STATUS "mbox1" (MESSAGES 1)`)
		c.S(`A009 OK (^_^)`)

		// Check that the message is still in mbox2
		// The message no longer has the recent flag.
		c.C(`A00A SELECT mbox2`)
		c.Se(`* 1 EXISTS`)
		c.Se(`A00A OK [READ-WRITE] (^_^)`)

		// Flag, unflag, expunge and check the message is still there.
		c.C(`A00B STORE 1 +FLAGS (\Deleted)`)
		c.S(`* 1 FETCH (FLAGS (\Deleted \Seen))`)
		c.Sx(`A00B OK .* command completed in .*`)
		c.C(`A00C STORE 1 -FLAGS (\Deleted)`)
		c.S(`* 1 FETCH (FLAGS (\Seen))`)
		c.Sx(`A00C OK .* command completed in .*`)
		c.C(`A00D EXPUNGE`)
		c.S(`A00D OK (^_^)`)
		c.C(`A00E STATUS mbox1 (MESSAGES)`)
		c.S(`* STATUS "mbox1" (MESSAGES 1)`)
		c.S(`A00E OK (^_^)`)
	})
}

func TestUIDDeleted(t *testing.T) {
	runOneToOneTestWithAuth(t, "user", "pass", "/", func(c *testConnection, _ *testSession) {
		// Create two mailboxes
		c.C("b001 CREATE mbox1")
		c.S("b001 OK (^_^)")
		c.C("b001 CREATE mbox2")
		c.S("b001 OK (^_^)")

		// Create a message in mbox1
		c.doAppend(`mbox1`, `To: 1@pm.me`, `\Seen`).expect("OK")
		c.doAppend(`mbox1`, `To: 2@pm.me`, `\Seen`).expect("OK")
		c.C(`A002 SELECT mbox1`)
		c.Se(`A002 OK [READ-WRITE] (^_^)`)

		// Copy message 2 to mbox2 and flag it as deleted in mbox 1
		c.C(`A003 UID COPY 2 mbox2`)
		c.Sx(`A003 OK .*`)
		c.C(`A004 UID STORE 2 +FLAGS (\Deleted)`)
		c.S(`* 2 FETCH (FLAGS (\Deleted \Recent \Seen) UID 2)`)
		c.Sx(`A004 OK .* command completed in .*`)

		// Check that the copy in mbox2 is does not have the flag \Deleted
		c.C(`A005 SELECT mbox2`)
		c.Se(`* 1 EXISTS`)
		c.Se(`A005 OK [READ-WRITE] (^_^)`)
		c.C(`A006 UID FETCH 1 (FLAGS)`)
		c.S(`* 1 FETCH (FLAGS (\Recent \Seen) UID 1)`)
		c.Sx(`A006 OK .* command completed in .*`)

		// Expunge the copy in mbox1
		c.C(`A007 SELECT mbox1`)
		c.Se(`* 2 EXISTS`)
		c.Se(`A007 OK [READ-WRITE] (^_^)`)
		c.C(`A008 EXPUNGE`)
		c.S(`* 2 EXPUNGE`)
		c.Sx(`A008 OK .*`)
		c.C(`A009 STATUS mbox1 (MESSAGES)`)
		c.S(`* STATUS "mbox1" (MESSAGES 1)`)
		c.S(`A009 OK (^_^)`)

		// Check that the message is still in mbox2
		c.C(`A00A SELECT mbox2`)
		c.Se(`* 1 EXISTS`)
		c.Se(`A00A OK [READ-WRITE] (^_^)`)

		// Flag,unflag, expunge and check the message is still there.
		c.C(`A00B UID STORE 1 +FLAGS (\Deleted)`)
		c.S(`* 1 FETCH (FLAGS (\Deleted \Seen) UID 1)`)
		c.Sx(`A00B OK .* command completed in .*`)
		c.C(`A00C UID STORE 1 -FLAGS (\Deleted)`)
		c.S(`* 1 FETCH (FLAGS (\Seen) UID 1)`)
		c.Sx(`A00C OK .* command completed in .*`)
		c.C(`A00D EXPUNGE`)
		c.S(`A00D OK (^_^)`)
		c.C(`A00E STATUS mbox1 (MESSAGES)`)
		c.S(`* STATUS "mbox1" (MESSAGES 1)`)
		c.S(`A00E OK (^_^)`)
	})
}
