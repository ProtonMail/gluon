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

		// Test
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
		c.S(
			`* LIST (\Noselect) "/" "~/Mail/foo"`,
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

func TestListRef(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t, withDelimiter(".")), func(c *testConnection, s *testSession) {
		c.C(`tag create some.thing.else.entirely`).OK(`tag`)

		// No ref - names interpreted like SELECT
		c.C(`tag list "" some`).S(`* LIST (\Unmarked) "." "some"`).OK(`tag`)
		c.C(`tag list "" some.thing`).S(`* LIST (\Unmarked) "." "some.thing"`).OK(`tag`)
		c.C(`tag list "" some.thing.else`).S(`* LIST (\Unmarked) "." "some.thing.else"`).OK(`tag`)
		c.C(`tag list "" some.thing.else.entirely`).S(`* LIST (\Unmarked) "." "some.thing.else.entirely"`).OK(`tag`)

		// Level 1 ref
		c.C(`tag list "some." thing`).S(`* LIST (\Unmarked) "." "some.thing"`).OK(`tag`)
		c.C(`tag list "some." thing.else`).S(`* LIST (\Unmarked) "." "some.thing.else"`).OK(`tag`)
		c.C(`tag list "some." thing.else.entirely`).S(`* LIST (\Unmarked) "." "some.thing.else.entirely"`).OK(`tag`)

		// Level 2 ref
		c.C(`tag list "some.thing." else`).S(`* LIST (\Unmarked) "." "some.thing.else"`).OK(`tag`)
		c.C(`tag list "some.thing." else.entirely`).S(`* LIST (\Unmarked) "." "some.thing.else.entirely"`).OK(`tag`)

		// Level 3 ref
		c.C(`tag list "some.thing.else." entirely`).S(`* LIST (\Unmarked) "." "some.thing.else.entirely"`).OK(`tag`)

		// Empty ref
		c.C(`tag list "" ""`).S(`* LIST (\Noselect) "." ""`).OK(`tag`)
		c.C(`tag list "some" ""`).S(`* LIST (\Noselect) "." ""`).OK(`tag`)
		c.C(`tag list "some." ""`).S(`* LIST (\Noselect) "." "some."`).OK(`tag`)
		c.C(`tag list "some.thing" ""`).S(`* LIST (\Noselect) "." "some."`).OK(`tag`)
		c.C(`tag list "some.thing." ""`).S(`* LIST (\Noselect) "." "some."`).OK(`tag`)
		c.C(`tag list "some.thing.else" ""`).S(`* LIST (\Noselect) "." "some."`).OK(`tag`)
		c.C(`tag list "some.thing.else." ""`).S(`* LIST (\Noselect) "." "some."`).OK(`tag`)
		c.C(`tag list "some.thing.else.entirely" ""`).S(`* LIST (\Noselect) "." "some."`).OK(`tag`)
		c.C(`tag list "some.thing.else.entirely." ""`).S(`* LIST (\Noselect) "." "some."`).OK(`tag`)
	})
}

func TestListInbox(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t, withDelimiter(".")), func(c *testConnection, s *testSession) {
		c.C(`tag create other`).OK(`tag`)
		c.C(`tag create inboxx`).OK(`tag`)
		c.C(`tag create inbox.child`).OK(`tag`)

		// Inbox is matched case-insensitively.
		c.C(`tag list "" "inbox"`).S(`* LIST (\Unmarked) "." "INBOX"`).OK(`tag`)
		c.C(`tag list "" "INBOX"`).S(`* LIST (\Unmarked) "." "INBOX"`).OK(`tag`)
		c.C(`tag list "" "iNbOx"`).S(`* LIST (\Unmarked) "." "INBOX"`).OK(`tag`)

		// Inbox is matched case-insensitively even when it's just part of a path.
		c.C(`tag list "" "inbox.child"`).S(`* LIST (\Unmarked) "." "INBOX.child"`).OK(`tag`)
		c.C(`tag list "" "INBOX.child"`).S(`* LIST (\Unmarked) "." "INBOX.child"`).OK(`tag`)
		c.C(`tag list "" "iNbOx.child"`).S(`* LIST (\Unmarked) "." "INBOX.child"`).OK(`tag`)

		// Inbox is matched case-insensitively when it's split over ref and pattern.
		c.C(`tag list "inb" "ox"`).S(`* LIST (\Unmarked) "." "INBOX"`).OK(`tag`)
		c.C(`tag list "INB" "ox"`).S(`* LIST (\Unmarked) "." "INBOX"`).OK(`tag`)
		c.C(`tag list "inb" "OX"`).S(`* LIST (\Unmarked) "." "INBOX"`).OK(`tag`)
		c.C(`tag list "INB" "OX"`).S(`* LIST (\Unmarked) "." "INBOX"`).OK(`tag`)

		// Other mailboxes are matched case-sensitively.
		c.C(`tag list "" "other"`).S(`* LIST (\Unmarked) "." "other"`).OK(`tag`)
		c.C(`tag list "" "OTHER"`).Sx(`OK`)
		c.C(`tag list "" "oThEr"`).Sx(`OK`)

		// Other mailboxes are matched case-sensitively.
		c.C(`tag list "" "inboxx"`).S(`* LIST (\Unmarked) "." "inboxx"`).OK(`tag`)
		c.C(`tag list "" "INBOXX"`).Sx(`OK`)
		c.C(`tag list "" "iNbOxX"`).Sx(`OK`)
	})
}

func TestListRemoved(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t, withDelimiter(".")), func(c *testConnection, s *testSession) {
		c.C(`tag create blurdybloop`).OK(`tag`)
		c.C(`tag create foo.bar`).OK(`tag`)

		c.C(`tag list "" "*"`).S(
			`* LIST (\Unmarked) "." "foo"`,
			`* LIST (\Unmarked) "." "foo.bar"`,
			`* LIST (\Unmarked) "." "blurdybloop"`,
			`* LIST (\Unmarked) "." "INBOX"`,
		).OK(`tag`)

		c.C(`tag delete blurdybloop`).OK(`tag`)
		c.C(`tag delete foo`).OK(`tag`)

		c.C(`tag list "" "*"`).S(
			`* LIST (\Noselect) "." "foo"`,
			`* LIST (\Unmarked) "." "foo.bar"`,
			`* LIST (\Unmarked) "." "INBOX"`,
		).OK(`tag`)

		c.C(`tag list "" "%"`).S(
			`* LIST (\Noselect) "." "foo"`,
			`* LIST (\Unmarked) "." "INBOX"`,
		).OK(`tag`)
	})
}

func TestListPanic(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t, withDelimiter(".")), func(c *testConnection, s *testSession) {
		s.mailboxCreated("user", []string{"no-parent", "just-child"})

		c.C(`S001 list "" "*"`)
		c.S(
			`* LIST (\Unmarked) "." "INBOX"`,
			`* LIST (\Noselect) "." "no-parent"`,
			`* LIST (\Unmarked) "." "no-parent.just-child"`,
		)
		c.OK(`S001`)

		c.C(`P001 list "" "%"`)
		c.S(
			`* LIST (\Unmarked) "." "INBOX"`,
			`* LIST (\Noselect) "." "no-parent"`,
		)
		c.OK(`P001`)
	})
}

func TestListWildcards(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t, withDelimiter(".")), func(c *testConnection, s *testSession) {
		c.C(`tag create some.thing.else.entirely`).OK(`tag`)
		c.C(`tag delete some`).OK(`tag`)

		c.C(`tag list "" "*"`)
		c.S(
			`* LIST (\Unmarked) "." "INBOX"`,
			`* LIST (\Noselect) "." "some"`,
			`* LIST (\Unmarked) "." "some.thing"`,
			`* LIST (\Unmarked) "." "some.thing.else"`,
			`* LIST (\Unmarked) "." "some.thing.else.entirely"`,
		)
		c.OK(`tag`)

		c.C(`tag list "" "%"`)
		c.S(
			`* LIST (\Unmarked) "." "INBOX"`,
			`* LIST (\Noselect) "." "some"`,
		)
		c.OK(`tag`)

		c.C(`tag list "some.thing" "*"`)
		c.S(
			`* LIST (\Unmarked) "." "some.thing"`,
			`* LIST (\Unmarked) "." "some.thing.else"`,
			`* LIST (\Unmarked) "." "some.thing.else.entirely"`,
		)
		c.OK(`tag`)

		c.C(`tag list "some.thing" "%"`)
		c.S(`* LIST (\Unmarked) "." "some.thing"`)
		c.OK(`tag`)

		c.C(`tag list "some.thing." "*"`)
		c.S(
			`* LIST (\Unmarked) "." "some.thing.else"`,
			`* LIST (\Unmarked) "." "some.thing.else.entirely"`,
		)
		c.OK(`tag`)

		c.C(`tag list "some.thing." "%"`)
		c.S(`* LIST (\Unmarked) "." "some.thing.else"`)
		c.OK(`tag`)
	})
}

func TestListSpecialUseAttributes(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t, withDelimiter(".")), func(c *testConnection, s *testSession) {
		s.mailboxCreated("user", []string{"INBOX"})
		s.mailboxCreatedWithAttributes("user", []string{"Koncepty"}, imap.NewFlagSet(imap.AttrDrafts))
		s.mailboxCreatedWithAttributes("user", []string{"Odeslane"}, imap.NewFlagSet(imap.AttrSent))
		s.mailboxCreatedWithAttributes("user", []string{"S hvezdickou"}, imap.NewFlagSet(imap.AttrFlagged))
		s.mailboxCreatedWithAttributes("user", []string{"Archiv"}, imap.NewFlagSet(imap.AttrArchive))
		s.mailboxCreatedWithAttributes("user", []string{"Spam"}, imap.NewFlagSet(imap.AttrJunk))
		s.mailboxCreatedWithAttributes("user", []string{"Kos"}, imap.NewFlagSet(imap.AttrTrash))
		s.mailboxCreatedWithAttributes("user", []string{"Vsechny zpravy"}, imap.NewFlagSet(imap.AttrAll))

		c.C(`a list "" "*"`)
		c.S(
			`* LIST (\Unmarked) "." "INBOX"`,
			`* LIST (\Drafts \Unmarked) "." "Koncepty"`,
			`* LIST (\Sent \Unmarked) "." "Odeslane"`,
			`* LIST (\Flagged \Unmarked) "." "S hvezdickou"`,
			`* LIST (\Archive \Unmarked) "." "Archiv"`,
			`* LIST (\Junk \Unmarked) "." "Spam"`,
			`* LIST (\Trash \Unmarked) "." "Kos"`,
			`* LIST (\All \Unmarked) "." "Vsechny zpravy"`,
		)
		c.OK(`a`)
	})
}
