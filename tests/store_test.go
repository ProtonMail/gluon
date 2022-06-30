package tests

import (
	"testing"
)

func TestStore(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C("b001 CREATE saved-messages")
		c.S("b001 OK (^_^)")

		c.doAppend(`saved-messages`, `To: 1@pm.me`, `\Seen`).expect("OK")
		c.doAppend(`saved-messages`, `To: 2@pm.me`).expect("OK")
		c.doAppend(`saved-messages`, `To: 3@pm.me`, `\Seen`).expect("OK")

		c.C(`A002 SELECT saved-messages`)
		c.Se(`A002 OK [READ-WRITE] (^_^)`)

		// TODO: Match flags in any order.
		c.C(`A005 FETCH 1:* (FLAGS)`)
		c.S(`* 1 FETCH (FLAGS (\Recent \Seen))`,
			`* 2 FETCH (FLAGS (\Recent))`,
			`* 3 FETCH (FLAGS (\Recent \Seen))`)
		c.Sx(`A005 OK .* command completed in .*`)

		// Add \Deleted and \Draft to the first message.
		c.C(`A003 STORE 1 +FLAGS (\Deleted \Draft)`)
		c.S(`* 1 FETCH (FLAGS (\Deleted \Draft \Recent \Seen))`)
		c.Sx(`A003 OK .* command completed in .*`)
		c.C(`A005 FETCH 1 (FLAGS)`)
		c.S(`* 1 FETCH (FLAGS (\Deleted \Draft \Recent \Seen))`)
		c.Sx(`A005 OK .* command completed in .*`)

		// Remove \Seen from the second message (which does not have it)
		c.C(`A003 STORE 2 -FLAGS (\Seen)`)
		c.Sx(`A003 OK .* command completed in .*`)
		c.C(`A005 FETCH 2 (FLAGS)`)
		c.S(`* 2 FETCH (FLAGS (\Recent))`)
		c.Sx(`A005 OK .* command completed in .*`)

		// Replace the third message's flags with \Flagged and \Answered.
		c.C(`A003 STORE 3 FLAGS (\Flagged \Answered)`)
		c.S(`* 3 FETCH (FLAGS (\Answered \Flagged \Recent))`)
		c.Sx(`A003 OK .* command completed in .*`)
		c.C(`A005 FETCH 3 (FLAGS)`)
		c.S(`* 3 FETCH (FLAGS (\Answered \Flagged \Recent))`)
		c.Sx(`A005 OK .* command completed in .*`)

		// Any attempt to alter the \Recent flags will fail
		c.C(`A003 STORE 3 FLAGS (\Recent Test)`).BAD(`A003`)
		c.C(`A003 STORE 3 +FLAGS (\RECENT Test)`).BAD(`A003`)
		c.C(`A003 STORE 3 -FLAGS (\RECent)`).BAD(`A003`)
	})
}

func TestStoreSilent(t *testing.T) {
	runManyToOneTestWithAuth(t, defaultServerOptions(t), []int{1, 2}, func(c map[int]*testConnection, s *testSession) {
		// one message in INBOX
		c[1].doAppend(`INBOX`, `To: 1@pm.me`).expect("OK")

		// 2 sessions with INBOX selected
		// Message is only recent in the first.
		for i := 1; i <= 2; i++ {
			c[i].C(`A001 SELECT INBOX`)
			c[i].Sxe(`A001 OK`)
		}

		// FLAGS. Both sessions get the untagged FETCH response
		c[1].C(`A002 STORE 1 FLAGS (flag1 flag2)`)
		c[1].S(`* 1 FETCH (FLAGS (\Recent flag1 flag2))`)
		c[1].Sx(`A002 OK`)
		c[2].C(`B002 NOOP`)
		c[2].S(`* 1 FETCH (FLAGS (flag1 flag2))`)
		c[2].Sx(`B002 OK`)

		// +FLAGS. Both sessions get the untagged FETCH response
		c[1].C(`A003 STORE 1 +FLAGS (flag3)`)
		c[1].S(`* 1 FETCH (FLAGS (\Recent flag1 flag2 flag3))`)
		c[1].Sx(`A003 OK`)
		c[2].C(`B003 NOOP`)
		c[2].S(`* 1 FETCH (FLAGS (flag1 flag2 flag3))`)
		c[2].Sx(`B003 OK`)

		// -FLAGS. Both sessions get the untagged FETCH response
		c[1].C(`A004 STORE 1 -FLAGS (flag3 flag2)`)
		c[1].S(`* 1 FETCH (FLAGS (\Recent flag1))`)
		c[1].Sx(`A004 OK`)
		c[2].C(`B004 NOOP`)
		c[2].S(`* 1 FETCH (FLAGS (flag1))`)
		c[2].Sx(`B004 OK`)

		// FLAGS.SILENT Only session 2 gets the untagged FETCH response
		c[1].C(`A005 STORE 1 FLAGS.SILENT (flag1 flag2)`)
		c[1].Sx(`A005 OK`)
		c[2].C(`B005 NOOP`)
		c[2].S(`* 1 FETCH (FLAGS (flag1 flag2))`)
		c[2].Sx(`B005 OK`)

		// +FLAGS.SILENT Only session 2 gets the untagged FETCH response
		c[1].C(`A006 STORE 1 +FLAGS.SILENT (flag3)`)
		c[1].Sx(`A006 OK`)
		c[2].C(`B006 NOOP`)
		c[2].S(`* 1 FETCH (FLAGS (flag1 flag2 flag3))`)
		c[2].Sx(`B006 OK`)

		// -FLAGS.SILENT Only session 2 gets the untagged FETCH response
		c[1].C(`A007 STORE 1 -FLAGS.SILENT (flag3 flag2)`)
		c[1].Sx(`A007 OK`)
		c[2].C(`B007 NOOP`)
		c[2].S(`* 1 FETCH (FLAGS (flag1))`)
		c[2].Sx(`B007 OK`)
	})
}

func TestUIDStore(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C("b001 CREATE saved-messages")
		c.S("b001 OK (^_^)")

		c.doAppend(`saved-messages`, `To: 1@pm.me`, `\Seen`).expect("OK")
		c.doAppend(`saved-messages`, `To: 2@pm.me`).expect("OK")
		c.doAppend(`saved-messages`, `To: 3@pm.me`, `\Seen`).expect("OK")

		c.C(`A002 SELECT saved-messages`)
		c.Se(`A002 OK [READ-WRITE] (^_^)`)

		// TODO: Match flags in any order.
		c.C(`A005 FETCH 1:* (FLAGS)`)
		c.S(`* 1 FETCH (FLAGS (\Recent \Seen))`,
			`* 2 FETCH (FLAGS (\Recent))`,
			`* 3 FETCH (FLAGS (\Recent \Seen))`)
		c.Sx(`A005 OK .* command completed in .*`)

		// Add \Deleted and \Draft to the first message.
		c.C(`A003 UID STORE 1 +FLAGS (\Deleted \Draft)`)
		c.S(`* 1 FETCH (FLAGS (\Deleted \Draft \Recent \Seen) UID 1)`)
		c.Sx(`A003 OK .* command completed in .*`)
		c.C(`A005 UID FETCH 1 (FLAGS)`)
		c.S(`* 1 FETCH (FLAGS (\Deleted \Draft \Recent \Seen) UID 1)`)
		c.Sx(`A005 OK .* command completed in .*`)

		// Remove \Seen from the second message.
		c.C(`A003 UID STORE 2 -FLAGS (\Seen)`)
		c.Sx(`A003 OK .* command completed in .*`)
		c.C(`A005 UID FETCH 2 (FLAGS)`)
		c.S(`* 2 FETCH (FLAGS (\Recent) UID 2)`)
		c.Sx(`A005 OK .* command completed in .*`)

		// Replace the third message's flags with \Flagged and \Answered.
		c.C(`A003 UID STORE 3 FLAGS (\Flagged \Answered)`)
		c.S(`* 3 FETCH (FLAGS (\Answered \Flagged \Recent) UID 3)`)
		c.Sx(`A003 OK .* command completed in .*`)
		c.C(`A005 UID FETCH 3 (FLAGS)`)
		c.S(`* 3 FETCH (FLAGS (\Answered \Flagged \Recent) UID 3)`)
		c.Sx(`A005 OK .* command completed in .*`)

		// Any attempt to alter the \Recent flags will fail
		c.C(`A003 UID STORE 3 FLAGS (\Recent Test)`).BAD(`A003`)
		c.C(`A003 UID STORE 3 +FLAGS (\RECENT Test)`).BAD(`A003`)
		c.C(`A003 UID STORE 3 -FLAGS (\RECent)`).BAD(`A003`)
	})
}

func TestFlagsDuplicateAndCaseInsensitive(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.doAppend(`INBOX`, `To: 1@pm.me`).expect("OK")

		c.C(`A001 SELECT INBOX`)
		c.Se(`A001 OK [READ-WRITE] (^_^)`)

		// no duplicates
		c.C(`A002 STORE 1 FLAGS (flag1 flag1)`)
		c.S(`* 1 FETCH (FLAGS (\Recent flag1))`)
		c.Sx(`A002 OK`)

		// no duplicates and case-insensitive
		c.C(`A003 STORE 1 FLAGS (FLAG1 flag2 FLAG2)`)
		c.S(`* 1 FETCH (FLAGS (FLAG1 \Recent flag2))`)
		c.Sx(`A003 OK`)

		// unchanged, no untagged response
		c.C(`A004 STORE 1 FLAGS (FLAG1 flag2 FLAG2)`)
		c.Sx(`A004 OK`)
		c.C(`A005 FETCH 1 (FLAGS)`)
		c.S(`* 1 FETCH (FLAGS (FLAG1 \Recent flag2))`)
		c.Sx(`A005 OK`)

		// +FLAGS with no changes, no untagged response
		c.C(`A006 STORE 1 +FLAGS (flag1 flag2)`)
		c.Sx(`A006 OK`)

		// +FLAGS with a new flag
		c.C(`A007 STORE 1 +FLAGS (FLAG1 FLAG3)`)
		c.S(`* 1 FETCH (FLAGS (FLAG1 FLAG3 \Recent flag2))`)
		c.Sx(`A007 OK`)

		// +FLAGS with a empty flag list, no untagged response
		c.C(`A008 STORE 1 +FLAGS ()`)
		c.Sx(`A008 OK`)

		// -FLAGS with difference case
		c.C(`A009 STORE 1 -FLAGS (flag3)`)
		c.S(`* 1 FETCH (FLAGS (FLAG1 \Recent flag2))`)
		c.Sx(`A009 OK`)

		// -FLAGS -with non-existing flag, no untagged
		c.C(`A00A STORE 1 -FLAGS (flag3)`)
		c.Sx(`A00A OK`)

		// -FLAGS with empty list
		c.C(`A00B STORE 1 -FLAGS ()`)
		c.Sx(`A00B OK`)
	})
}
