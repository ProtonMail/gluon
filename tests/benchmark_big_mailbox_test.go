package tests

import (
	"fmt"
	"testing"

	goimap "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/stretchr/testify/require"
)

func BenchmarkBigMailboxStatus(b *testing.B) {
	runOneToOneTestWithAuth(b, defaultServerOptions(b), func(c *testConnection, s *testSession) {
		mboxID := s.mailboxCreated("user", []string{"mbox"})

		ids := s.batchMessageCreated("user", mboxID, 32515, func(n int) ([]byte, []string) {
			return []byte(fmt.Sprintf(`To: %v@pm.me`, n)), []string{}
		})

		b.Run("status", func(b *testing.B) {
			c.Cf(`A001 STATUS %v (MESSAGES)`, "mbox").Sx(fmt.Sprintf(`MESSAGES %v`, len(ids))).OK(`A001`)
		})
	})
}

func BenchmarkBigMailboxFetchSequence(b *testing.B) {
	runOneToOneTestClientWithAuth(b, defaultServerOptions(b), func(client *client.Client, s *testSession) {
		mboxID := s.mailboxCreated("user", []string{"mbox"})

		ids := s.batchMessageCreated("user", mboxID, 128515, func(n int) ([]byte, []string) {
			return []byte(fmt.Sprintf(`To: %v@pm.me`, n)), []string{}
		})

		_, err := client.Select("mbox", false)
		require.NoError(b, err)

		fetchResult := newFetchCommand(b, client).withItems(goimap.FetchAll).fetch("1:*")

		if len(fetchResult.messages) != len(ids) {
			panic("Fetch failed")
		}
	})
}

func BenchmarkBigMailboxFetchUID(b *testing.B) {
	runOneToOneTestClientWithAuth(b, defaultServerOptions(b), func(client *client.Client, s *testSession) {
		mboxID := s.mailboxCreated("user", []string{"mbox"})

		ids := s.batchMessageCreated("user", mboxID, 128515, func(n int) ([]byte, []string) {
			return []byte(fmt.Sprintf(`To: %v@pm.me`, n)), []string{}
		})

		_, err := client.Select("mbox", false)
		require.NoError(b, err)

		fetchResult := newFetchCommand(b, client).withItems(goimap.FetchAll).fetchUid("1:*")

		if len(fetchResult.messages) != len(ids) {
			panic("Fetch failed")
		}
	})
}

func BenchmarkListManyMailboxes(b *testing.B) {
	runOneToOneTestClientWithAuth(b, defaultServerOptions(b), func(client *client.Client, s *testSession) {
		s.batchMailboxCreated("user", 64515, func(i int) string { return fmt.Sprintf("mbox_%v", i) })

		listMailboxesClient(b, client, "", "*")
	})
}
