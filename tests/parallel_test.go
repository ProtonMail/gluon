package tests

import (
	"math/rand"
	"testing"

	"github.com/ProtonMail/gluon/imap"
	"github.com/bradenaw/juniper/xslices"
	"github.com/emersion/go-imap/client"
	"github.com/stretchr/testify/require"
)

func TestSelectWhileSyncing(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, s *testSession) {
		// Define some mailbox names.
		mailboxNames := []string{"Apply", "B", "C", "D", "E", "F", "G", "H", "I", "J"}

		// Collect the mailbox IDs as they're created.
		mailboxIDs := xslices.Map(mailboxNames, func(name string) imap.LabelID {
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
			mboxName := mailboxNames[rand.Int()%len(mailboxNames)] //nolint:gosec
			_, err := client.Select(mboxName, false)
			require.NoError(t, err)

			_, err = client.Select("INBOX", false)
			require.NoError(t, err)
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
		c.Se("A006 OK [READ-WRITE] SELECT")

		// Do some fetches in parallel.
		c.C(`A005 FETCH 1:200 (BODY[TEXT])`)
		c.C(`A006 FETCH 201:400 (BODY[TEXT])`)
		c.C(`A007 FETCH 401:600 (BODY[TEXT])`)
		c.C(`A008 FETCH 601:800 (BODY[TEXT])`)
		c.C(`A009 FETCH 801:1000 (BODY[TEXT])`)

		// The fetch commands will complete eventually; who knows which will be processed first.
		// TODO: Also check the untagged FETCH responses.
		c.Sxe(
			`A005 OK command completed in .*`,
			`A006 OK command completed in .*`,
			`A007 OK command completed in .*`,
			`A008 OK command completed in .*`,
			`A009 OK command completed in .*`,
		)

		// We should then be able to logout fine.
		c.C("A010 logout")
		c.S("* BYE")
		c.OK("A010")

		// Logging out should close the connection.
		c.expectClosed()
	})
}
