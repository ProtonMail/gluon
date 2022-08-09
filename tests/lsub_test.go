package tests

import (
	"testing"
)

func TestLsub(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t, withDelimiter(".")), func(c *testConnection, _ *testSession) {
		c.C("B002 CREATE #news.comp.mail.mime")
		c.OK("B002")

		c.C(`A003 UNSUBSCRIBE "#news.comp.mail.mime"`)
		c.OK("A003")

		c.C("B003 CREATE #news.comp.mail.misc")
		c.OK("B003")

		c.C(`A003 UNSUBSCRIBE "#news.comp.mail.misc"`)
		c.OK("A003")

		c.C(`A002 LSUB "#news." "comp.mail.*"`)
		c.OK("A002")

		c.C(`A003 SUBSCRIBE "#news.comp.mail.mime"`)
		c.OK("A003")

		c.C(`A004 LSUB "#news." "comp.mail.*"`)
		c.S(`* LSUB (\Unmarked) "." "#news.comp.mail.mime"`)
		c.OK("A004")

		c.C(`A005 SUBSCRIBE "#news.comp.mail.misc"`)
		c.OK("A005")

		c.C(`A006 LSUB "#news." "comp.mail.*"`)
		c.S(`* LSUB (\Unmarked) "." "#news.comp.mail.mime"`,
			`* LSUB (\Unmarked) "." "#news.comp.mail.misc"`)
		c.OK(`A006`)

		// TODO(GODT-1612): Handle this edge case properly.
		// c.C(`A007 LSUB "#news." "comp.%"`)
		// c.S(`* LSUB (\Noselect) "." "#news.comp.mail"`)
		// c.OK(`A007`)
	})
}
