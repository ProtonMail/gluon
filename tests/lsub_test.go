package tests

import (
	"testing"

	"github.com/ProtonMail/gluon/imap"
)

func TestLsub(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t, withDelimiter(".")), func(c *testConnection, _ *testSession) {
		c.C("tag CREATE foo.bar").OK(`tag`)
		c.C(`tag UNSUBSCRIBE foo`).OK(`tag`)

		c.C(`S001 LSUB "" "*"`)
		c.S(
			`* LSUB (\Unmarked) "." "INBOX"`,
			`* LSUB (\Unmarked) "." "foo.bar"`,
		)
		c.S("S001 OK LSUB")

		c.C(`P001 LSUB "" "%"`)
		c.S(
			`* LSUB (\Unmarked) "." "INBOX"`,
			`* LSUB (\Noselect) "." "foo"`,
		)
		c.S("P001 OK LSUB")

		c.C(`tag DELETE foo`).OK(`tag`)

		c.C(`S101 LSUB "" "*"`)
		c.S(
			`* LSUB (\Unmarked) "." "INBOX"`,
			`* LSUB (\Unmarked) "." "foo.bar"`,
		)
		c.S("S101 OK LSUB")

		c.C(`P101 LSUB "" "%"`)
		c.S(
			`* LSUB (\Unmarked) "." "INBOX"`,
			`* LSUB (\Noselect) "." "foo"`,
		)
		c.S("P101 OK LSUB")

		c.C("tag CREATE #news.comp.mail.mime").OK("tag")

		c.C(`tag UNSUBSCRIBE "#news.comp.mail.mime"`).OK("tag")

		c.C("tag CREATE #news.comp.mail.misc").OK("tag")

		c.C(`tag UNSUBSCRIBE "#news.comp.mail.misc"`).OK("tag")

		c.C(`S002 LSUB "#news." "comp.mail.*"`)
		c.S("S002 OK LSUB")

		c.C(`P002 LSUB "#news." "comp.mail.%"`)
		c.S("P002 OK LSUB")

		c.C(`tag SUBSCRIBE "#news.comp.mail.mime"`).OK("tag")

		c.C(`S003 LSUB "#news." "comp.mail.*"`)
		c.S(`* LSUB (\Unmarked) "." "#news.comp.mail.mime"`)
		c.S("S003 OK LSUB")

		c.C(`P003 LSUB "#news." "comp.mail.%"`)
		c.S(`* LSUB (\Unmarked) "." "#news.comp.mail.mime"`)
		c.S("P003 OK LSUB")

		c.C(`tag SUBSCRIBE "#news.comp.mail.misc"`).OK("tag")

		c.C(`S004 LSUB "#news." "comp.mail.*"`)
		c.S(
			`* LSUB (\Unmarked) "." "#news.comp.mail.mime"`,
			`* LSUB (\Unmarked) "." "#news.comp.mail.misc"`,
		)
		c.OK(`S004`)

		c.C(`P004 LSUB "#news." "comp.mail.%"`)
		c.S(
			`* LSUB (\Unmarked) "." "#news.comp.mail.mime"`,
			`* LSUB (\Unmarked) "." "#news.comp.mail.misc"`,
		)
		c.OK(`P004`)

		c.C(`S005 LSUB "#news." "comp.*"`)
		c.S(
			`* LSUB (\Unmarked) "." "#news.comp.mail"`,
			`* LSUB (\Unmarked) "." "#news.comp.mail.mime"`,
			`* LSUB (\Unmarked) "." "#news.comp.mail.misc"`,
		)
		c.OK(`S005`)

		c.C(`P005 LSUB "#news." "comp.%"`)
		c.S(`* LSUB (\Unmarked) "." "#news.comp.mail"`)
		c.OK(`P005`)
	})
}

func TestLsubPanic(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t, withDelimiter(".")), func(c *testConnection, s *testSession) {
		s.mailboxCreated("user", []string{"no-parent", "just-child"})

		c.C(`S001 LSUB "" "*"`)
		c.S(
			`* LSUB (\Unmarked) "." "INBOX"`,
			`* LSUB (\Unmarked) "." "no-parent.just-child"`,
		)
		c.OK(`S001`)

		c.C(`P001 LSUB "" "%"`)
		c.S(
			`* LSUB (\Unmarked) "." "INBOX"`,
			`* LSUB (\Noselect) "." "no-parent"`,
		)
		c.OK(`P001`)
	})
}

func TestLsubSubscribedNotExisting(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t, withDelimiter(".")), func(c *testConnection, _ *testSession) {
		c.C("tag CREATE foo").OK(`tag`)

		c.C(`A001 LSUB "" "foo"`)
		c.S(`* LSUB (\Unmarked) "." "foo"`)
		c.OK(`A001`)

		c.C(`tag DELETE foo`).OK(`tag`)

		c.C(`A002 LSUB "" "foo"`)
		// TODO GODT-1896: The server MUST NOT unilaterally remove an existing mailbox name
		// from the subscription list even if a mailbox by that name no
		// longer exists.
		//
		// c.S(`* LSUB (\Noselect) "." "foo"`)
		c.OK(`A002`)
	})
}

func TestLsubWithHiddenMailbox(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t, withDelimiter(".")), func(c *testConnection, s *testSession) {
		m1 := s.mailboxCreatedWithAttributes("user", []string{"Koncepty"}, imap.NewFlagSet())
		s.mailboxCreatedWithAttributes("user", []string{"Odeslane"}, imap.NewFlagSet())
		m2 := s.mailboxCreatedWithAttributes("user", []string{"S hvezdickou"}, imap.NewFlagSet())
		s.mailboxCreatedWithAttributes("user", []string{"Archiv"}, imap.NewFlagSet())
		m3 := s.mailboxCreatedWithAttributes("user", []string{"Spam"}, imap.NewFlagSet())
		s.mailboxCreatedWithAttributes("user", []string{"Kos"}, imap.NewFlagSet())
		m4 := s.mailboxCreatedWithAttributes("user", []string{"Vsechny zpravy"}, imap.NewFlagSet())

		{
			connector := s.conns[s.userIDs["user"]]
			connector.SetMailboxVisible(m1, false)
			connector.SetMailboxVisible(m2, false)
			connector.SetMailboxVisible(m3, false)
			connector.SetMailboxVisible(m4, false)
		}

		{
			c.C("A001 UNSUBSCRIBE Koncepty").OK("A001")
			c.C("A001 UNSUBSCRIBE \"Vsechny zpravy\"").OK("A001")
			c.C("A001 UNSUBSCRIBE Archiv").OK("A001")
			c.C("A001 UNSUBSCRIBE Kos").OK("A001")
		}

		c.C(`a LSUB "" "*"`)
		c.S(
			`* LSUB (\Unmarked) "." "INBOX"`,
			`* LSUB (\Unmarked) "." "Odeslane"`,
		)
		c.OK(`a`)
	})
}

func TestLSubWithUtf8MailboxNames(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t, withDelimiter(".")), func(c *testConnection, s *testSession) {
		s.mailboxCreated("user", []string{"mbox-öüäëæøå"})
		s.flush("user")
		c.C(`a LSUB "" "*"`)
		c.S(
			`* LSUB (\Unmarked) "." "INBOX"`,
			`* LSUB (\Unmarked) "." "mbox-&APYA,ADkAOsA5gD4AOU-"`,
		)
		c.OK(`a`)
	})
}
