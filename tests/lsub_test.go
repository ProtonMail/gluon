package tests

import (
	"testing"
)

func TestLsub(t *testing.T) {
	runOneToOneTestWithAuth(t, "user", "pass", ".", func(c *testConnection, _ *testSession) {
		c.C("B002 CREATE #news.comp.mail.mime")
		c.S("B002 OK (^_^)")

		c.C(`A003 UNSUBSCRIBE "#news.comp.mail.mime"`)
		c.S(`A003 OK (^_^)`)

		c.C("B003 CREATE #news.comp.mail.misc")
		c.S("B003 OK (^_^)")

		c.C(`A003 UNSUBSCRIBE "#news.comp.mail.misc"`)
		c.S(`A003 OK (^_^)`)

		c.C(`A002 LSUB "#news." "comp.mail.*"`)
		c.S(`A002 OK (^_^)`)

		c.C(`A003 SUBSCRIBE "#news.comp.mail.mime"`)
		c.S(`A003 OK (^_^)`)

		c.C(`A004 LSUB "#news." "comp.mail.*"`)
		c.S(`* LSUB (\Unmarked) "." "#news.comp.mail.mime"`)
		c.S(`A004 OK (^_^)`)

		c.C(`A005 SUBSCRIBE "#news.comp.mail.misc"`)
		c.S(`A005 OK (^_^)`)

		c.C(`A006 LSUB "#news." "comp.mail.*"`)
		c.S(`* LSUB (\Unmarked) "." "#news.comp.mail.mime"`,
			`* LSUB (\Unmarked) "." "#news.comp.mail.misc"`)
		c.S(`A006 OK (^_^)`)

		c.C(`A007 LSUB "#news." "comp.%"`)
		c.S(`* LSUB (\NoSelect) "." "#news.comp.mail"`)
		c.S(`A007 OK (^_^)`)
	})
}
