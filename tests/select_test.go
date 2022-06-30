package tests

import (
	"testing"
)

func TestSelect(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C("A002 CREATE Archive")
		c.S("A002 OK (^_^)")

		c.doAppend(`INBOX`, `To: 1@pm.me`, `\Seen`).expect("OK")
		c.doAppend(`INBOX`, `To: 2@pm.me`).expect("OK")
		c.doAppend(`Archive`, `To: 3@pm.me`, `\Seen`).expect("OK")

		c.C("A006 select INBOX")
		c.S(`* FLAGS (\Deleted \Flagged \Seen)`,
			`* 2 EXISTS`,
			`* 2 RECENT`,
			`* OK [UNSEEN 2]`,
			`* OK [PERMANENTFLAGS (\Deleted \Flagged \Seen)]`,
			`* OK [UIDNEXT 3]`,
			`* OK [UIDVALIDITY 1]`)
		c.S("A006 OK [READ-WRITE] (^_^)")

		// Selecting again modifies the RECENT value.
		c.C("A006 select INBOX")
		c.S(`* FLAGS (\Deleted \Flagged \Seen)`,
			`* 2 EXISTS`,
			`* 0 RECENT`,
			`* OK [UNSEEN 2]`,
			`* OK [PERMANENTFLAGS (\Deleted \Flagged \Seen)]`,
			`* OK [UIDNEXT 3]`,
			`* OK [UIDVALIDITY 1]`)
		c.S("A006 OK [READ-WRITE] (^_^)")

		c.C("A007 select Archive")
		c.S(`* FLAGS (\Deleted \Flagged \Seen)`,
			`* 1 EXISTS`,
			`* 1 RECENT`,
			`* OK [PERMANENTFLAGS (\Deleted \Flagged \Seen)]`,
			`* OK [UIDNEXT 2]`,
			`* OK [UIDVALIDITY 1]`)
		c.S(`A007 OK [READ-WRITE] (^_^)`)
	})
}

func TestSelectNoSuchMailbox(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C("a003 select What")
		c.Sx("a003 NO .*")
	})
}

func TestSelectUTF7(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		// Test we can create a mailbox with a UTF-7 name.
		c.C("A003 CREATE &ZeVnLIqe-").OK("A003")

		// Test we can select the mailbox.
		c.C("A004 SELECT &ZeVnLIqe-").OK("A004")

		// The mailbox should appear in LIST responses with the same UTF-7 encoding.
		c.C(`A005 LIST "" "*"`).Sxe(`&ZeVnLIqe-`).OK(`A005`)
	})
}
