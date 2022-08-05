package tests

import (
	"testing"
)

func TestSequenceRange(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C("a001 CREATE mbox1")
		c.S("a001 OK CREATE")
		c.C("a002 CREATE mbox2")
		c.S("a002 OK CREATE")
		c.C(`A003 SELECT mbox1`)
		c.Se(`A003 OK [READ-WRITE] SELECT`)

		// return BAD for any sequence range in an empty mailbox
		c.C(`A004 FETCH 1 (FLAGS)`).BAD(`A004`)
		c.C(`A005 FETCH * (FLAGS)`).BAD(`A005`)
		c.C(`A006 FETCH 1:* (FLAGS)`).BAD(`A006`)

		c.doAppend(`mbox1`, `To: 1@pm.me`).expect("OK")
		c.doAppend(`mbox1`, `To: 2@pm.me`).expect("OK")
		c.doAppend(`mbox1`, `To: 3@pm.me`).expect("OK")
		c.doAppend(`mbox1`, `To: 4@pm.me`).expect("OK")
		c.doAppend(`mbox1`, `To: 5@pm.me`).expect("OK")

		// test various set of ranges with STORE, FETCH, MOVE & COPY
		c.C(`A007 FETCH 1 (FLAGS)`)
		c.Sx(`\* 1 FETCH`)
		c.OK(`A007`)
		c.C(`A008 FETCH 6 (FLAGS)`).BAD(`A008`)
		c.C(`A009 FETCH 1,3:4 (FLAGS)`)
		for i := 0; i < 3; i++ {
			c.Sx(`\* \d FETCH`)
		}
		c.OK(`A009`)
		c.C(`A010 STORE 1,2,3,4 +FLAGS (flag)`)
		for i := 0; i < 4; i++ {
			c.Sx(`\* \d FETCH \(FLAGS \(\\Recent flag\)\)`)
		}
		c.OK(`A010`)
		c.C(`A011 COPY 1,3:* mbox2`)
		c.S(`A011`)
		c.C(`A012 COPY 6:* mbox2`).BAD(`A012`)
		c.C(`A012 COPY 6:* mbox2`).BAD(`A012`)
		c.C(`A013 MOVE 1,5,3 mbox2`)
		c.Sx(`\* OK \[COPYUID`)
		for i := 0; i < 3; i++ {
			c.Sx(`\* \d EXPUNGE`)
		}
		c.OK(`A013`)

		// test ranges given in reverse order
		c.doAppend(`mbox1`, `To: 6@pm.me`).expect("OK")
		c.doAppend(`mbox1`, `To: 7@pm.me`).expect("OK")
		c.C(`A014 STORE 4:2 -FLAGS (flag)`)
		c.Sx(`\* 2 FETCH `) // 2 was the only message in 4:2 to have flag set
		c.OK(`A014`)
		c.C(`A015 COPY *:1 mbox2`)
		c.OK(`A015`)
		c.C(`A016 COPY 7:5 mbox2`).BAD(`A016`)
		c.C(`A017 COPY 4:3,1,2:1 mbox2`)
		c.OK(`A017`)
	})
}

func TestUIDSequenceRange(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		// UID based operation will send an OK response for any grammatically valid UID sequence set
		// if no message match the UID sequence set, the operations simply return OK with no untagged response before.

		c.C("a001 CREATE mbox1")
		c.S("a001 OK CREATE")
		c.C("a002 CREATE mbox2")
		c.S("a002 OK CREATE")
		c.C(`A003 SELECT mbox1`)
		c.Se(`A003 OK [READ-WRITE] SELECT`)

		// return OK for any UID sequence range in an empty mailbox
		c.C(`A004 UID FETCH 1 (FLAGS)`)
		c.OK(`A004`)
		c.C(`A005 UID FETCH * (FLAGS)`)
		c.OK(`A005`)
		c.C(`A006 UID FETCH 1:* (FLAGS)`)
		c.OK(`A006`)

		c.doAppend(`mbox1`, `To: 1@pm.me`).expect("OK")
		c.doAppend(`mbox1`, `To: 2@pm.me`).expect("OK")
		c.doAppend(`mbox1`, `To: 3@pm.me`).expect("OK")
		c.doAppend(`mbox1`, `To: 4@pm.me`).expect("OK")
		c.doAppend(`mbox1`, `To: 5@pm.me`).expect("OK")

		//// test various set of ranges with STORE, FETCH, MOVE & COPY
		c.C(`A007 UID FETCH 1 (FLAGS)`)
		c.Sx(`\* 1 FETCH`)
		c.OK(`A007`)
		c.C(`A008 UID FETCH 6 (FLAGS)`)
		c.OK(`A008`)
		c.C(`A009 UID FETCH 1,3:4 (FLAGS)`)
		for i := 0; i < 3; i++ {
			c.Sx(`\* \d FETCH .*`)
		}
		c.OK(`A009`)
		c.C(`A010 UID STORE 1,2,3,4 +FLAGS (flag)`)
		for i := 0; i < 4; i++ {
			c.Sx(`\* \d FETCH \(FLAGS \(\\Recent flag\) UID \d\)`)
		}
		c.OK(`A010`)
		c.C(`A011 UID COPY 1,3:* mbox2`)
		c.OK(`A011`)
		c.C(`A012 UID COPY 6:* mbox2`)
		c.OK(`A012`)
		c.C(`A012 UID COPY 1:* mbox2`)
		c.Sx(`A012 OK \[COPYUID`)
		c.C(`A013 UID MOVE 1,5,3 mbox2`)
		c.Sx(`\* OK \[COPYUID`)
		for i := 0; i < 3; i++ {
			c.Sx(`\* \d EXPUNGE`)
		}
		c.OK(`A013`)

		// test ranges given in reverse order
		c.doAppend(`mbox1`, `To: 6@pm.me`).expect("OK")
		c.doAppend(`mbox1`, `To: 7@pm.me`).expect("OK")
		c.C(`A014 UID STORE 4:2 -FLAGS (flag)`)
		for i := 0; i < 2; i++ {
			c.Sx(`\* \d FETCH`)
		}
		c.OK(`A014`)
		c.C(`A015 UID COPY *:1 mbox2`)
		c.OK(`A015`)
		c.C(`A016 UID COPY 7:5 mbox2`)
		c.OK(`A016`)
		c.C(`A017 UID COPY 4:3,1,2:1 mbox2`)
		c.OK(`A017`)
	})
}

func TestWildcard(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		// Create an empty mailbox.
		c.C("tag create mbox").OK("tag")
		c.C("tag select mbox").OK("tag")

		// FETCH with wildcard returns BAD.
		c.C("tag fetch * (flags)").BAD("tag")

		// UID FETCH with wildcard returns OK.
		c.C("tag uid fetch * (flags)").OK("tag")
	})
}
