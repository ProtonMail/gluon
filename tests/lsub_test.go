package tests

import (
	"testing"
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
