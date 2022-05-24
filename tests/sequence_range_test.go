package tests

import (
	"testing"
)

func TestSequenceRange(t *testing.T) {
	runOneToOneTestWithAuth(t, "user", "pass", "/", func(c *testConnection, _ *testSession) {
		c.C("a001 CREATE mbox1")
		c.S("a001 OK (^_^)")
		c.C("a002 CREATE mbox2")
		c.S("a002 OK (^_^)")
		c.C(`A003 SELECT mbox1`)
		c.Se(`A003 OK [READ-WRITE] (^_^)`)

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
		c.Sx(`A007 OK .*`)
		c.C(`A008 FETCH 6 (FLAGS)`).BAD(`A008`)
		c.C(`A009 FETCH 1,3:4 (FLAGS)`)
		for i := 0; i < 3; i++ {
			c.Sx(`\* \d FETCH`)
		}
		c.Sx(`A009 OK`)
		c.C(`A010 STORE 1,2,3,4 +FLAGS (flag)`)
		for i := 0; i < 4; i++ {
			c.Sx(`\* \d FETCH \(FLAGS \(\\Recent flag\)\)`)
		}
		c.Sx(`A010 OK .* command completed in`)
		c.C(`A011 COPY 1,3:* mbox2`)
		c.Sx(`A011 OK .* command completed in`)
		c.C(`A012 COPY 6:* mbox2`).BAD(`A012`)
		c.C(`A012 COPY 6:* mbox2`).BAD(`A012`)
		c.C(`A013 MOVE 1,5,3 mbox2`)
		c.Sx(`\* OK \[COPYUID`)
		for i := 0; i < 3; i++ {
			c.Sx(`\* \d EXPUNGE`)
		}
		c.Sx(`A013 OK .* command completed in`)

		// test ranges given in reverse order
		c.doAppend(`mbox1`, `To: 6@pm.me`).expect("OK")
		c.doAppend(`mbox1`, `To: 7@pm.me`).expect("OK")
		c.C(`A014 STORE 4:2 -FLAGS (flag)`)
		c.Sx(`\* 2 FETCH `) // 2 was the only message in 4:2 to have flag set
		c.Sx(`A014 OK .* command completed in`)
		c.C(`A015 COPY *:1 mbox2`)
		c.Sx(`A015 OK`)
		c.C(`A016 COPY 7:5 mbox2`).BAD(`A016`)
		c.C(`A017 COPY 4:3,1,2:1 mbox2`)
		c.Sx(`A017 OK`)
	})
}

func TestUIDSequenceRange(t *testing.T) {
	runOneToOneTestWithAuth(t, "user", "pass", "/", func(c *testConnection, _ *testSession) {
		// UID based operation will send an OK response for any grammatically valid UID sequence set
		// if no message match the UID sequence set, the operations simply return OK with no untagged response before.

		c.C("a001 CREATE mbox1")
		c.S("a001 OK (^_^)")
		c.C("a002 CREATE mbox2")
		c.S("a002 OK (^_^)")
		c.C(`A003 SELECT mbox1`)
		c.Se(`A003 OK [READ-WRITE] (^_^)`)

		// return OK for any UID sequence range in an empty mailbox
		c.C(`A004 UID FETCH 1 (FLAGS)`)
		c.Sx(`A004 OK`)
		c.C(`A005 UID FETCH * (FLAGS)`)
		c.Sx(`A005 OK`)
		c.C(`A006 UID FETCH 1:* (FLAGS)`)
		c.Sx(`A006 OK`)

		c.doAppend(`mbox1`, `To: 1@pm.me`).expect("OK")
		c.doAppend(`mbox1`, `To: 2@pm.me`).expect("OK")
		c.doAppend(`mbox1`, `To: 3@pm.me`).expect("OK")
		c.doAppend(`mbox1`, `To: 4@pm.me`).expect("OK")
		c.doAppend(`mbox1`, `To: 5@pm.me`).expect("OK")

		//// test various set of ranges with STORE, FETCH, MOVE & COPY
		c.C(`A007 UID FETCH 1 (FLAGS)`)
		c.Sx(`\* 1 FETCH`)
		c.Sx(`A007 OK`)
		c.C(`A008 UID FETCH 6 (FLAGS)`)
		c.Sx(`A008 OK`)
		c.C(`A009 UID FETCH 1,3:4 (FLAGS)`)
		for i := 0; i < 3; i++ {
			c.Sx(`\* \d FETCH .*`)
		}
		c.Sx(`A009 OK .*`)
		c.C(`A010 UID STORE 1,2,3,4 +FLAGS (flag)`)
		for i := 0; i < 4; i++ {
			c.Sx(`\* \d FETCH \(FLAGS \(\\Recent flag\) UID \d\)`)
		}
		c.Sx(`A010 OK .* command completed in`)
		c.C(`A011 UID COPY 1,3:* mbox2`)
		c.Sx(`A011 OK .* command completed in`)
		c.C(`A012 UID COPY 6:* mbox2`)
		c.Sx(`A012 OK`)
		c.C(`A012 UID COPY 1:* mbox2`)
		c.Sx(`A012 OK \[COPYUID`)
		c.C(`A013 UID MOVE 1,5,3 mbox2`)
		c.Sx(`\* OK \[COPYUID`)
		for i := 0; i < 3; i++ {
			c.Sx(`\* \d EXPUNGE`)
		}
		c.Sx(`A013 OK .* command completed in`)

		// test ranges given in reverse order
		c.doAppend(`mbox1`, `To: 6@pm.me`).expect("OK")
		c.doAppend(`mbox1`, `To: 7@pm.me`).expect("OK")
		c.C(`A014 UID STORE 4:2 -FLAGS (flag)`)
		for i := 0; i < 2; i++ {
			c.Sx(`\* \d FETCH`)
		}
		c.Sx(`A014 OK .* command completed in`)
		c.C(`A015 UID COPY *:1 mbox2`)
		c.Sx(`A015 OK`)
		c.C(`A016 UID COPY 7:5 mbox2`)
		c.Sx(`A016 OK`)
		c.C(`A017 UID COPY 4:3,1,2:1 mbox2`)
		c.Sx(`A017 OK`)
	})
}
