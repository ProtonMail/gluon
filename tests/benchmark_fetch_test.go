package tests

import (
	"strconv"
	"testing"
)

func BenchmarkFetchDatabase(b *testing.B) {
	for n := 0; n < 15; n++ {
		mboxSize := 1 << n

		b.Run(strconv.Itoa(mboxSize), func(b *testing.B) {
			runOneToOneTestWithAuth(b, defaultServerOptions(b), func(c *testConnection, s *testSession) {
				benchID := s.mailboxCreated("user", []string{"BENCH"})

				for i := 0; i < mboxSize; i++ {
					s.messageCreatedFromFile("user", benchID, `testdata/multipart-mixed.eml`)
				}

				c.C(`Apply SELECT BENCH`)
				c.Se(`Apply OK [READ-WRITE] (^_^)`)

				c.doBench(b, `FETCH * (ENVELOPE)`)
				c.doBench(b, `FETCH 1:* (ENVELOPE)`)

				c.doBench(b, `FETCH * (FLAGS)`)
				c.doBench(b, `FETCH 1:* (FLAGS)`)

				c.doBench(b, `FETCH * (INTERNALDATE)`)
				c.doBench(b, `FETCH 1:* (INTERNALDATE)`)

				c.doBench(b, `FETCH * (BODY)`)
				c.doBench(b, `FETCH 1:* (BODY)`)

				c.doBench(b, `FETCH * (BODYSTRUCTURE)`)
				c.doBench(b, `FETCH 1:* (BODYSTRUCTURE)`)

				c.doBench(b, `FETCH * (RFC822.SIZE)`)
				c.doBench(b, `FETCH 1:* (RFC822.SIZE)`)

				c.doBench(b, `FETCH * (UID)`)
				c.doBench(b, `FETCH 1:* (UID)`)
			})
		})
	}
}

func BenchmarkFetchSingleCache(b *testing.B) {
	for n := 0; n < 15; n++ {
		mboxSize := 1 << n

		b.Run(strconv.Itoa(mboxSize), func(b *testing.B) {
			runOneToOneTestWithAuth(b, defaultServerOptions(b), func(c *testConnection, s *testSession) {
				benchID := s.mailboxCreated("user", []string{"BENCH"})

				for i := 0; i < mboxSize; i++ {
					s.messageCreatedFromFile("user", benchID, `testdata/multipart-mixed.eml`)
				}

				c.C(`Apply SELECT BENCH`)
				c.Se(`Apply OK [READ-WRITE] (^_^)`)

				c.doBench(b, `FETCH * (RFC822)`)
				c.doBench(b, `FETCH 1:* (RFC822)`)

				c.doBench(b, `FETCH * (RFC822.HEADER)`)
				c.doBench(b, `FETCH 1:* (RFC822.HEADER)`)

				c.doBench(b, `FETCH * (RFC822.TEXT)`)
				c.doBench(b, `FETCH 1:* (RFC822.TEXT)`)

				c.doBench(b, `FETCH * (BODY[HEADER])`)
				c.doBench(b, `FETCH 1:* (BODY[HEADER])`)

				c.doBench(b, `FETCH * (BODY[HEADER.FIELDS (To From Date)])`)
				c.doBench(b, `FETCH 1:* (BODY[HEADER.FIELDS (To From Date)])`)

				c.doBench(b, `FETCH * (BODY[HEADER.FIELDS.NOT (To From Date)])`)
				c.doBench(b, `FETCH 1:* (BODY[HEADER.FIELDS.NOT (To From Date)])`)

				c.doBench(b, `FETCH * (BODY[TEXT])`)
				c.doBench(b, `FETCH 1:* (BODY[TEXT])`)

				c.doBench(b, `FETCH * (BODY[])`)
				c.doBench(b, `FETCH 1:* (BODY[])`)

				c.doBench(b, `FETCH * (BODY[1])`)
				c.doBench(b, `FETCH 1:* (BODY[1])`)

				c.doBench(b, `FETCH * (BODY[2])`)
				c.doBench(b, `FETCH 1:* (BODY[2])`)

				c.doBench(b, `FETCH * (BODY[1.MIME])`)
				c.doBench(b, `FETCH 1:* (BODY[1.MIME])`)

				c.doBench(b, `FETCH * (BODY[2.MIME])`)
				c.doBench(b, `FETCH 1:* (BODY[2.MIME])`)
			})
		})
	}
}
