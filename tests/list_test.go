package tests

import (
	"testing"

	"github.com/ProtonMail/gluon/imap"
)

func TestList(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C("A002 CREATE #news/comp/mail/mime")
		c.S("A002 OK (^_^)")

		c.C("A003 CREATE /usr/staff/jones")
		c.S("A003 OK (^_^)")

		c.C("A004 CREATE ~/Mail/meetings")
		c.S("A004 OK (^_^)")

		c.C("A005 CREATE ~/Mail/foo/bar")
		c.S("A005 OK (^_^)")

		// Delete the parent, leaving the child behind.
		// The deleted parent will be reported with \Noselect.
		c.C("A005 DELETE ~/Mail/foo")
		c.S("A005 OK (^_^)")

		c.C(`A101 LIST "" ""`)
		c.S(`* LIST (\Noselect) "/" ""`)
		c.S(`A101 OK (^_^)`)

		c.C(`A102 LIST #news/comp/mail/misc ""`)
		c.S(`* LIST (\Noselect) "/" "#news/"`)
		c.S(`A102 OK (^_^)`)

		c.C(`A103 LIST /usr/staff/jones ""`)
		c.S(`* LIST (\Noselect) "/" "/"`)
		c.S(`A103 OK (^_^)`)

		c.C(`A202 LIST ~/Mail/ %`)
		c.S(`* LIST (\Noselect) "/" "~/Mail/foo"`,
			`* LIST (\Unmarked) "/" "~/Mail/meetings"`)
		c.S(`A202 OK (^_^)`)
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
		c.S(`A103 OK (^_^)`)

		s.messageCreatedFromFile("user", mailboxID, "testdata/multipart-mixed.eml")

		c.C(`A103 LIST "" *`)
		c.S(`* LIST (\Unmarked) "/" "INBOX"`,
			`* LIST (\Marked \Noinferiors) "/" "custom-attributes"`)
		c.S(`A103 OK (^_^)`)
	})
}
