package tests

import (
	"fmt"
	"testing"

	"github.com/ProtonMail/gluon/imap"
)

func TestMove(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox, mboxID string) {
		// There are 100 messages in the origin and no messages in the destination.
		c.Cf(`A001 status %v (messages)`, mbox).Sxe(`MESSAGES 100`).OK(`A001`)
		c.C(`A002 status inbox (messages)`).Sxe(`MESSAGES 0`).OK(`A002`)

		// Move half the messages to the destination.
		c.C(`A003 move 1:50 inbox`)
		c.Sx(`OK \[COPYUID 1 1:50 1:50\]`)
		c.Sx(repeat(`\d EXPUNGE`, 50)...)
		c.OK(`A003`)

		// There are now 50 messages in the origin and 50 messages in the destination.
		c.Cf(`A004 status %v (messages)`, mbox).Sxe(`MESSAGES 50`).OK(`A004`)
		c.C(`A005 status inbox (messages)`).Sxe(`MESSAGES 50`).OK(`A005`)

		// Move the other half the messages to the destination (this time using UID MOVE).
		c.C(`A006 uid move 51:100 inbox`)
		c.Sx(`OK \[COPYUID 1 51:100 51:100\]`)
		c.Sx(repeat(`\d EXPUNGE`, 50)...)
		c.OK(`A006`)

		// There are now no messages in the origin and 100 messages in the destination.
		c.Cf(`A007 status %v (messages)`, mbox).Sxe(`MESSAGES 0`).OK(`A007`)
		c.C(`A008 status inbox (messages)`).Sxe(`MESSAGES 100`).OK(`A008`)
	})
}

func TestMoveTryCreate(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox, mboxID string) {
		// There are 100 messages in the origin.
		c.Cf(`A001 status %v (messages)`, mbox).Sxe(`MESSAGES 100`).OK(`A001`)

		// MOVE to a nonexistent destination.
		c.C(`A002 move 1:* this-name-does-not-exist`)
		c.Sx(`A002 NO \[TRYCREATE\]`)

		// UID MOVE to a nonexistent destination.
		c.C(`A002 uid move 1:* this-name-does-not-exist`)
		c.Sx(`A002 NO \[TRYCREATE\]`)
	})
}

func TestMoveNonExistent(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox, mboxID string) {
		// MOVE some of the messages out of the mailbox.
		c.C(`A001 move 1:24,76:100 inbox`).OK(`A001`)

		// Attempting to MOVE nonexistent messages by sequence number returns BAD.
		c.C(`A002 move 51:100 inbox`).BAD(`A002`)

		// Attempting to UID MOVE nonexistent messages with UIDs lower than the smallest in the mailbox returns OK.
		c.C(`A003 uid move 1:24 inbox`)
		// Nothing should be returned
		c.Sx(`A003 OK .*`)

		// Attempting to UID MOVE nonexistent messages with UIDs higher than the largest in the mailbox returns OK.
		c.C(`A004 uid move 76:100 inbox`)
		// Nothing should be returned
		c.Sx(`A004 OK .*`)

		c.C(`A005 uid move 24:26 inbox`)
		c.S("* OK [COPYUID 1 25:26 50:51]")
		c.OK(`A005`)
	})
}

func _TestMoveBackAndForthRepeated(t *testing.T) {
	for i := 1; i <= 1000; i++ {
		fmt.Printf("Iteration %d\n", i)
		TestMoveBackAndForth(t)
	}
}

func TestMoveBackAndForth(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox, mboxID string) {
		// There are 100 messages in the origin and no messages in the destination.
		c.Cf(`A001 status %v (messages)`, mbox).Sxe(`MESSAGES 100`).OK(`A001`)
		c.C(`A002 status inbox (messages)`).Sxe(`MESSAGES 0`).OK(`A002`)

		// Move the messages to the destination.
		c.C(`A003 move 1:* inbox`).OK(`A003`)

		// There are now 100 messages in the destination, none in the origin.
		c.Cf(`A004 status %v (messages)`, mbox).Sxe(`MESSAGES 0`).OK(`A004`)
		c.C(`A005 status inbox (messages)`).Sxe(`MESSAGES 100`).OK(`A005`)

		// Move the messages back to the origin.
		c.C(`A006 select inbox`).OK(`A006`)
		c.Cf(`A007 uid move 1:* %v`, mbox).OK(`A007`)

		// There are 100 messages in the origin and no messages in the destination.
		c.Cf(`A008 status %v (messages)`, mbox).Sxe(`MESSAGES 100`).OK(`A008`)
		c.C(`A009 status inbox (messages)`).Sxe(`MESSAGES 0`).OK(`A009`)
	})
}

func TestMoveCopyDuplicates(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		// 4 messages in inbox.
		c.doAppend("inbox", "To: 1@pm.me").expect("OK")
		c.doAppend("inbox", "To: 2@pm.me").expect("OK")
		c.doAppend("inbox", "To: 3@pm.me").expect("OK")
		c.doAppend("inbox", "To: 4@pm.me").expect("OK")

		// Create other mailbox.
		c.C("tag create other").OK("tag")

		// Move all from inbox to other (inbox: 0, other: 4)
		c.C("tag select inbox").OK("tag")
		c.C("tag move 1:* other").OK("tag")
		c.C("tag status inbox (messages)").Sxe("MESSAGES 0").OK("tag")
		c.C("tag status other (messages)").Sxe("MESSAGES 4").OK("tag")

		// Move all from other to other (inbox: 0, other: 4)
		c.C("tag select other").OK("tag")
		c.C("tag move 1:* other").OK("tag")
		c.C("tag status inbox (messages)").Sxe("MESSAGES 0").OK("tag")
		c.C("tag status other (messages)").Sxe("MESSAGES 4").OK("tag")

		// Copy all from other to other (inbox: 0, other: 4)
		c.C("tag copy 1:* other").OK("tag")
		c.C("tag status inbox (messages)").Sxe("MESSAGES 0").OK("tag")
		c.C("tag status other (messages)").Sxe("MESSAGES 4").OK("tag")

		// Copy first half from other to inbox (inbox: 2, other: 4)
		c.C("tag copy 1:2 inbox").OK("tag")
		c.C("tag status inbox (messages)").Sxe("MESSAGES 2").OK("tag")
		c.C("tag status other (messages)").Sxe("MESSAGES 4").OK("tag")

		// Move second half from other to inbox (inbox: 4, other: 2)
		c.C("tag move 3:4 inbox").OK("tag")
		c.C("tag status inbox (messages)").Sx("MESSAGES 4").OK("tag")
		c.C("tag status other (messages)").Sx("MESSAGES 2").OK("tag")

		// Move first half from other to inbox (inbox: 4, other: 0)
		c.C("tag move 1:2 inbox").OK("tag")
		c.C("tag status inbox (messages)").Sx("MESSAGES 4").OK("tag")
		c.C("tag status other (messages)").Sx("MESSAGES 0").OK("tag")
	})
}

func TestMoveDuplicate(t *testing.T) {
	runManyToOneTestWithAuth(t, defaultServerOptions(t), []int{1, 2, 3}, func(c map[int]*testConnection, s *testSession) {
		origID := s.mailboxCreated("user", []string{"orig"})
		destID := s.mailboxCreated("user", []string{"dest"})

		// Put three messages in the origin.
		for idx := 1; idx <= 3; idx++ {
			s.messageCreatedFromFile("user", origID, "testdata/text-plain.eml", imap.FlagSeen)
		}

		// Put three messages in the destination.
		for idx := 1; idx <= 3; idx++ {
			s.messageCreatedFromFile("user", destID, "testdata/text-plain.eml", imap.FlagSeen)
		}

		// Initially there are three messages in the origin.
		c[1].C(`A001 select orig`).OK(`A001`)
		c[3].C(`C001 status orig (messages)`).Sxe(`MESSAGES 3`).OK(`C001`)

		// Initially there are three messages in the destination.
		c[2].C(`B001 select dest`).OK(`B001`)
		c[3].C(`C002 status dest (messages)`).Sxe(`MESSAGES 3`).OK(`C002`)
		c[2].C(`B002 idle`).S("+ Ready")

		// Copy three messages into the destination.
		// They receive UIDs 4:6 in the destination.
		// Expect to receive exists updates for them.
		// The destination should then contain 6 messages.
		c[1].C(`A002 copy 1:3 dest`).OK(`A002`, `COPYUID 1 1:3 4:6`)
		c[2].Se(`* 4 EXISTS`).Se(`* 5 EXISTS`).Se(`* 6 EXISTS`)
		c[3].C(`C003 status dest (messages)`).Sxe(`MESSAGES 6`).OK(`C003`)

		// The origin still has three messages.
		c[3].C(`C004 status orig (messages)`).Sxe(`MESSAGES 3`).OK(`C004`)

		// Copy the same three messages into the destination.
		// The existing messages are removed and the new ones are added afterwards.
		// They receive UIDs 7:9 in the destination.
		// The destination should still contains six messages.
		c[1].C(`A002 copy 1:3 dest`).OK(`A002`, `COPYUID 1 1:3 7:9`)
		c[3].C(`C003 status dest (messages)`).Sxe(`MESSAGES 6`).OK(`C003`)

		// The origin still has three messages.
		c[3].C(`C004 status orig (messages)`).Sxe(`MESSAGES 3`).OK(`C004`)

		// Move the same three messages to the destination.
		// The existing messages are removed and the new ones are added afterwards.
		// They receive UIDs 10:12 in the destination.
		// The destination should still contains six messages.
		c[1].C(`A003 move 1:3 dest`).Sxe(`COPYUID 1 1:3 10:12`).OK(`A003`)
		c[3].C(`C003 status dest (messages)`).Sxe(`MESSAGES 6`).OK(`C003`)

		// The origin now is now empty.
		c[3].C(`C004 status orig (messages)`).Sxe(`MESSAGES 0`).OK(`C004`)

		// Finish idle.
		c[2].C(`DONE`).OK(`B002`)
	})
}
