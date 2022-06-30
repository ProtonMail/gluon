package tests

import (
	"math/rand"
	"testing"

	"github.com/bradenaw/juniper/xslices"
)

func TestSelectWhileSyncing(t *testing.T) {
	runManyToOneTestWithAuth(t, defaultServerOptions(t), []int{1, 2}, func(c map[int]*testConnection, s *testSession) {
		// Define some mailbox names.
		mailboxNames := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J"}

		// Collect the mailbox IDs as they're created.
		mailboxIDs := xslices.Map(mailboxNames, func(name string) string {
			return s.mailboxCreated("user", []string{name})
		})

		stopCh := make(chan struct{})
		doneCh := make(chan struct{})

		// Append a bunch of messages.
		go func() {
			defer close(doneCh)

			for {
				select {
				case <-stopCh:
					return

				default:
					s.messageCreatedFromFile("user", mailboxIDs[0], `testdata/multipart-mixed.eml`)
				}
			}
		}()

		// Select a bunch of mailboxes.
		for i := 0; i < 100; i++ {
			c[2].C("A006 select " + mailboxNames[rand.Int()%len(mailboxNames)]) //nolint:gosec
			c[2].Se("A006 OK [READ-WRITE] (^_^)")

			c[2].C("A006 select INBOX")
			c[2].Se("A006 OK [READ-WRITE] (^_^)")
		}

		// Stop appending.
		close(stopCh)

		// Wait for appending to finish.
		<-doneCh
	})
}

func TestTwoFetchesAtOnce(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		mboxID := s.mailboxCreated("user", []string{"mbox"})

		for i := 0; i < 1000; i++ {
			s.messageCreatedFromFile("user", mboxID, `testdata/multipart-mixed.eml`, `\Seen`)
		}

		c.C("A006 select mbox")
		c.Se("A006 OK [READ-WRITE] (^_^)")

		// Do some fetches in parallel.
		c.C(`A005 FETCH 1:200 (BODY[TEXT])`)
		c.C(`A006 FETCH 201:400 (BODY[TEXT])`)
		c.C(`A007 FETCH 401:600 (BODY[TEXT])`)
		c.C(`A008 FETCH 601:800 (BODY[TEXT])`)
		c.C(`A009 FETCH 801:1000 (BODY[TEXT])`)

		// The fetch commands will complete eventually; who knows which will be processed first.
		// TODO: Also check the untagged FETCH responses.
		c.Sxe(
			`A005 OK .* command completed in .*`,
			`A006 OK .* command completed in .*`,
			`A007 OK .* command completed in .*`,
			`A008 OK .* command completed in .*`,
			`A009 OK .* command completed in .*`,
		)

		// We should then be able to logout fine.
		c.C("A010 logout")
		c.S("* BYE (^_^)/~")
		c.S("A010 OK (^_^)")

		// Logging out should close the connection.
		c.expectClosed()
	})
}
