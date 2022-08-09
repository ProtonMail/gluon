package tests

import (
	"testing"
)

// GOMSRV-39: We should be able to match INBOX in other cases!
func TestMailboxCase(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C(`tag CREATE Archive`).OK(`tag`)
		c.C(`tag CREATE inbox/other`).OK(`tag`)

		// We can select INBOX in any case.
		c.C(`tag SELECT INBOX`).OK(`tag`)
		c.C(`tag SELECT inbox`).OK(`tag`)
		c.C(`tag SELECT iNbOx`).OK(`tag`)
		c.C(`tag SELECT INBOX/other`).OK(`tag`)
		c.C(`tag SELECT inbox/other`).OK(`tag`)
		c.C(`tag SELECT iNbOx/other`).OK(`tag`)

		// We can status INBOX in any case.
		c.C(`tag status INBOX (messages)`).OK(`tag`)
		c.C(`tag status inbox (messages)`).OK(`tag`)
		c.C(`tag status iNbOx (messages)`).OK(`tag`)
		c.C(`tag status INBOX/other (messages)`).OK(`tag`)
		c.C(`tag status inbox/other (messages)`).OK(`tag`)
		c.C(`tag status iNbOx/other (messages)`).OK(`tag`)

		// We can append INBOX in any case.
		c.C(`tag append INBOX () {11}`).Continue().C(`To: 1@pm.me`).OK(`tag`)
		c.C(`tag append inbox () {11}`).Continue().C(`To: 1@pm.me`).OK(`tag`)
		c.C(`tag append iNbOx () {11}`).Continue().C(`To: 1@pm.me`).OK(`tag`)
		c.C(`tag append INBOX/other () {11}`).Continue().C(`To: 1@pm.me`).OK(`tag`)
		c.C(`tag append inbox/other () {11}`).Continue().C(`To: 1@pm.me`).OK(`tag`)
		c.C(`tag append iNbOx/other () {11}`).Continue().C(`To: 1@pm.me`).OK(`tag`)

		// We can list inbox in any case.
		c.C(`tag LIST "" "INBOX"`).Sx(`INBOX`).OK(`tag`)
		c.C(`tag LIST "" "inbox"`).Sx(`INBOX`).OK(`tag`)
		c.C(`tag LIST "" "iNbOx"`).Sx(`INBOX`).OK(`tag`)
		c.C(`tag LIST "" "INBOX/other"`).Sx(`INBOX/other`).OK(`tag`)
		c.C(`tag LIST "" "inbox/other"`).Sx(`INBOX/other`).OK(`tag`)
		c.C(`tag LIST "" "iNbOx/other"`).Sx(`INBOX/other`).OK(`tag`)

		// We can only select non-inbox mailboxes in the original case.
		c.C(`tag SELECT Archive`).OK(`tag`)
		c.C(`tag SELECT ARCHIVE`).NO(`tag`)
		c.C(`tag SELECT archive`).NO(`tag`)
		c.C(`tag SELECT ArChIvE`).NO(`tag`)

		// We can only list non-inbox mailboxes in the original case.
		c.C(`tag LIST "" "Archive"`).Sx(`Archive`).OK(`tag`)
		c.C(`tag LIST "" "ARCHIVE"`).Sx(`tag OK`)
	})
}
