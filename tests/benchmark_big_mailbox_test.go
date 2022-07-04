package tests

import (
	"fmt"
	"testing"
)

func BenchmarkBigMailboxStatus(b *testing.B) {
	b.Run("status", func(b *testing.B) {
		runOneToOneTestWithAuth(b, defaultServerOptions(b), func(c *testConnection, s *testSession) {
			mboxID := s.mailboxCreated("user", []string{"mbox"})

			ids := s.batchMessageCreated("user", mboxID, 32515, func(n int) ([]byte, []string) {
				return []byte(fmt.Sprintf(`To: %v@pm.me`, n)), []string{}
			})

			c.Cf(`A001 STATUS %v (MESSAGES)`, "mbox").Sx(fmt.Sprintf(`MESSAGES %v`, len(ids))).OK(`A001`)
		})
	})
}
