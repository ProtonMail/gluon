package tests

import (
	"testing"
)

// GOMSRV-39: We should be able to match INBOX in other cases!
func _TestMailboxCase(t *testing.T) {
	runOneToOneTestWithAuth(t, "user", "pass", "/", func(c *testConnection, _ *testSession) {
		// Create a non-inbox mailbox.
		c.C(`A001 CREATE Archive`).OK(`A001`)

		// We can select INBOX in any case.
		c.C(`A002 SELECT INBOX`).OK(`A002`)
		c.C(`A003 SELECT inbox`).OK(`A003`)
		c.C(`A004 SELECT iNbOx`).OK(`A004`)

		// We can list inbox in any case.
		c.C(`A005 LIST "" "INBOX"`).Sx(`INBOX`).OK(`A005`)
		c.C(`A005 LIST "" "inbox"`).Sx(`INBOX`).OK(`A005`)

		// We can only select non-inbox mailboxes in the original case.
		c.C(`A004 SELECT Archive`).OK(`A004`)
		c.C(`A005 SELECT ARCHIVE`).NO(`A005`)
		c.C(`A006 SELECT archive`).NO(`A006`)
		c.C(`A007 SELECT ArChIvE`).NO(`A007`)

		// We can only list non-inbox mailboxes in the original case.
		c.C(`A005 LIST "" "Archive"`).Sx(`Archive`).OK(`A005`)
		c.C(`A005 LIST "" "ARCHIVE"`).Sx(`A005 OK`)
	})
}
