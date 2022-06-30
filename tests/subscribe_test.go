package tests

import (
	"testing"
)

func TestSubscribe(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t, withDelimiter(".")), func(c *testConnection, _ *testSession) {
		c.C("A002 CREATE #news.comp.mail.mime")
		c.S("A002 OK (^_^)")

		c.C("A003 SUBSCRIBE #this.name.does.not.exist")
		c.S("A003 NO no such mailbox (~_~)")

		// Mailboxes are subscribed by default.
		c.C("A004 UNSUBSCRIBE #news.comp.mail.mime")
		c.S("A004 OK (^_^)")

		c.C("A004 SUBSCRIBE #news.comp.mail.mime")
		c.S("A004 OK (^_^)")

		c.C("A005 SUBSCRIBE #news.comp.mail.mime")
		c.S("A005 NO already subscribed to this mailbox (~_~)")
	})
}
