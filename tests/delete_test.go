package tests

import (
	"testing"

	"github.com/ProtonMail/gluon/internal/response"
	goimap "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/stretchr/testify/require"
)

func TestDelete(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t, withDelimiter(".")), func(client *client.Client, _ *testSession) {
		require.NoError(t, client.Create("blurdybloop"))
		require.NoError(t, client.Create("foo"))
		require.NoError(t, client.Create("foo.bar"))

		checkMailboxesMatchNamesAndAttributes(
			t, client, "", "*",
			map[string][]string{
				"INBOX":       {goimap.UnmarkedAttr},
				"blurdybloop": {goimap.UnmarkedAttr},
				"foo":         {goimap.UnmarkedAttr},
				"foo.bar":     {goimap.UnmarkedAttr},
			},
		)

		require.NoError(t, client.Delete("blurdybloop"))
		require.NoError(t, client.Delete("foo"))

		checkMailboxesMatchNamesAndAttributes(
			t, client, "", "*",
			map[string][]string{
				"INBOX":   {goimap.UnmarkedAttr},
				"foo":     {goimap.NoSelectAttr},
				"foo.bar": {goimap.UnmarkedAttr},
			},
		)

		checkMailboxesMatchNamesAndAttributes(
			t, client, "", "%",
			map[string][]string{
				"INBOX": {goimap.UnmarkedAttr},
				"foo":   {goimap.NoSelectAttr},
			},
		)

		// deleting mailboxes with \Noselect Attribute is an error
		require.Error(t, client.Delete("foo"))
	})
}

func TestDeleteMailboxHasChildren(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C(`A001 CREATE "blurdybloop"`).OK(`A001`)
		c.C(`A002 CREATE "foo"`).OK(`A002`)
		c.C(`A003 CREATE "foo/bar"`).OK(`A003`)

		c.doAppend(`foo`, `To: 1@pm.me`).expect("OK")
		c.doAppend(`foo/bar`, `To: 2@pm.me`).expect("OK")

		c.C(`A004 SELECT "INBOX"`).OK(`A004`)
		c.C(`A005 LIST "" *`)
		c.S(
			`* LIST (\Unmarked) "/" "INBOX"`,
			`* LIST (\Unmarked) "/" "blurdybloop"`,
			`* LIST (\Marked) "/" "foo"`,
			`* LIST (\Marked) "/" "foo/bar"`,
		)
		c.OK(`A005`)

		c.C(`A006 STATUS "foo/bar" (UIDNEXT MESSAGES)`)
		c.S(`* STATUS "foo/bar" (UIDNEXT 2 MESSAGES 1)`)
		c.OK(`A006`)

		c.C(`A007 DELETE "foo"`).OK(`A007`)

		c.C(`A008 STATUS "foo/bar" (UIDNEXT MESSAGES)`)
		c.S(`* STATUS "foo/bar" (UIDNEXT 2 MESSAGES 1)`)
		c.OK(`A008`)

		c.C(`A009 LIST "" *`)
		c.S(
			`* LIST (\Unmarked) "/" "INBOX"`,
			`* LIST (\Unmarked) "/" "blurdybloop"`,
			`* LIST (\Noselect) "/" "foo"`,
			`* LIST (\Marked) "/" "foo/bar"`,
		)
		c.OK(`A009`)
	})
}

func TestDeleteCannotDeleteInbox(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		require.Error(t, client.Delete("INBOX"))
	})
}

func TestDeleteCannotDeleteMissingMailbox(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		require.Error(t, client.Delete("this doesn't exist"))
	})
}

func TestDeleteSelectedMailboxCausesDisconnect(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C("b001 CREATE mbox1")
		c.S("b001 OK CREATE")

		c.C("b002 SELECT mbox1").OK("b002")
		c.C("b003 DELETE mbox1").OK("b003")
		c.S(response.Bye().WithMailboxDeleted().String())
	})
}

func TestDeleteExaminedMailboxCausesDisconnect(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C("b001 CREATE mbox1")
		c.S("b001 OK CREATE")

		c.C("b002 EXAMINE mbox1").OK("b002")
		c.C("b003 DELETE mbox1").OK("b003")
		c.S(response.Bye().WithMailboxDeleted().String())
	})
}

func TestDeleteSelectedMailboxCausesDisconnectOnOtherClients(t *testing.T) {
	runManyToOneTestWithAuth(t, defaultServerOptions(t), []int{1, 2}, func(c map[int]*testConnection, s *testSession) {
		c[1].C("b001 CREATE mbox1")
		c[1].S("b001 OK CREATE")

		s.flush("user")

		c[2].C("c001 SELECT mbox1").OK("c001")

		// Delete mailbox
		c[1].C("b002 SELECT mbox1").OK("b002")
		c[1].C("b003 DELETE mbox1").OK("b003")
		c[1].S(response.Bye().WithMailboxDeleted().String())

		// Other client should get kicked out on next command
		c[2].C("c002 NOOP")
		c[2].S(response.Bye().WithInconsistentState().String())
	})
}

func TestDeleteExaminedMailboxCausesDisconnectOnOtherClients(t *testing.T) {
	runManyToOneTestWithAuth(t, defaultServerOptions(t), []int{1, 2}, func(c map[int]*testConnection, s *testSession) {
		c[1].C("b001 CREATE mbox1")
		c[1].S("b001 OK CREATE")

		s.flush("user")

		c[2].C("c001 EXAMINE mbox1").OK("c001")

		// Delete mailbox
		c[1].C("b002 EXAMINE mbox1").OK("b002")
		c[1].C("b003 DELETE mbox1").OK("b003")
		c[1].S(response.Bye().WithMailboxDeleted().String())

		// Other client should get kicked out on next command
		c[2].C("c002 NOOP")
		c[2].S(response.Bye().WithInconsistentState().String())
	})
}

func TestDeleteSelectedMailboxWithRemoteUpdateCausesDisconnect(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		mailboxID := s.mailboxCreated("user", []string{"mbox1"})
		s.flush("user")

		c.C("b002 SELECT mbox1").OK("b002")

		s.mailboxDeleted("user", mailboxID)
		s.flush("user")

		c.C("b003 NOOP")
		c.S(response.Bye().WithInconsistentState().String())
	})
}
