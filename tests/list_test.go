package tests

import (
	"testing"

	"github.com/ProtonMail/gluon/imap"
)

func TestList(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C("A002 CREATE #news/comp/mail/mime")
		c.OK("A002")

		c.C("A003 CREATE /usr/staff/jones")
		c.OK("A003")

		c.C("A004 CREATE ~/Mail/meetings")
		c.OK("A004")

		c.C("A005 CREATE ~/Mail/foo/bar")
		c.OK("A005")

		// Delete the parent, leaving the child behind.
		// The deleted parent will be reported with \Noselect.
		c.C("A005 DELETE ~/Mail/foo")
		c.OK("A005")

		c.C(`A101 LIST "" ""`)
		c.S(`* LIST (\Noselect) "/" ""`)
		c.OK("A101")

		c.C(`A102 LIST #news/comp/mail/misc ""`)
		c.S(`* LIST (\Noselect) "/" "#news/"`)
		c.OK("A102")

		c.C(`A103 LIST /usr/staff/jones ""`)
		c.S(`* LIST (\Noselect) "/" "/"`)
		c.OK("A103")

		c.C(`A202 LIST ~/Mail/ %`)
		c.S(`* LIST (\Noselect) "/" "~/Mail/foo"`,
			`* LIST (\Unmarked) "/" "~/Mail/meetings"`)
		c.OK("A202")
	})
}

func TestListFlagsAndAttributes(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		mailboxID := s.mailboxCreatedCustom(
			"user",
			[]string{"custom-attributes"},
			defaultFlags,
			defaultPermanentFlags,
			imap.NewFlagSet(imap.AttrNoInferiors),
		)

		c.C(`A103 LIST "" *`)
		c.S(`* LIST (\Unmarked) "/" "INBOX"`,
			`* LIST (\Noinferiors \Unmarked) "/" "custom-attributes"`)
		c.OK(`A103`)

		s.messageCreatedFromFile("user", mailboxID, "testdata/multipart-mixed.eml")

		c.C(`A103 LIST "" *`)
		c.S(`* LIST (\Unmarked) "/" "INBOX"`,
			`* LIST (\Marked \Noinferiors) "/" "custom-attributes"`)
		c.OK(`A103`)
	})
}
