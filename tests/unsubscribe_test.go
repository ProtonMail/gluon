package tests

import (
	"testing"
)

func TestUnsubscribe(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t, withDelimiter(".")), func(c *testConnection, _ *testSession) {
		c.C("A002 CREATE #news.comp.mail.mime")
		c.S("A002 OK CREATE")

		c.C("A003 UNSUBSCRIBE #this.name.does.not.exist")
		c.S("A003 NO no such mailbox")

		c.C("A005 UNSUBSCRIBE #news.comp.mail.mime")
		c.S("A005 OK UNSUBSCRIBE")

		c.C("A006 UNSUBSCRIBE #news.comp.mail.mime")
		c.S("A006 NO not subscribed to this mailbox")
	})
}
