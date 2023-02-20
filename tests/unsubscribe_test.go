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

func TestUnsubscribeAfterMailboxDeleted(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t, withDelimiter(".")), func(c *testConnection, _ *testSession) {
		c.C("A002 CREATE #news.comp.mail.mime")
		c.S("A002 OK CREATE")

		c.C("A006 DELETE #news.comp.mail.mime")
		c.S("A006 OK DELETE")

		c.C("A007 SUBSCRIBE #news.comp.mail.mime")
		c.S("A007 NO no such mailbox")

		c.C("A008 UNSUBSCRIBE #news.comp.mail.mime")
		c.S("A008 OK UNSUBSCRIBE")
	})
}

func TestUnsubscribeAfterMailboxRenamedDeleted(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t, withDelimiter(".")), func(c *testConnection, _ *testSession) {
		c.C("A002 CREATE mailbox")
		c.S("A002 OK CREATE")

		c.C("A002 RENAME mailbox mailbox2")
		c.S("A002 OK RENAME")

		c.C("A006 DELETE mailbox2")
		c.S("A006 OK DELETE")

		c.C("A008 UNSUBSCRIBE mailbox2")
		c.S("A008 OK UNSUBSCRIBE")
	})
}

func TestUnsubscribeList(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t, withDelimiter(".")), func(c *testConnection, _ *testSession) {
		c.C(`tag list "" "*"`)
		c.S(`* LIST (\Unmarked) "." "INBOX"`)
		c.OK(`tag`)

		c.C(`tag unsubscribe "INBOX"`).OK(`tag`)

		c.C(`tag list "" "*"`)
		c.S(`* LIST (\Unmarked) "." "INBOX"`)
		c.OK(`tag`)

		c.C(`tag lsub "" "*"`)
		c.OK(`tag`)

		c.C(`tag subscribe "INBOX"`).OK(`tag`)

		c.C(`tag lsub "" "*"`)
		c.S(`* LSUB (\Unmarked) "." "INBOX"`)
		c.OK(`tag`)

		c.C(`tag list "" "*"`)
		c.S(`* LIST (\Unmarked) "." "INBOX"`)
		c.OK(`tag`)
	})
}
