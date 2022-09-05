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
		{
			expectedMailboxNames := []string{
				"INBOX",
				"blurdybloop",
				"foo",
				"foo.bar",
			}
			expectedAttributes := []string{goimap.UnmarkedAttr}
			checkMailboxesMatchNamesAndAttributes(t, client, "", "*", expectedMailboxNames, expectedAttributes)
		}
		require.NoError(t, client.Delete("blurdybloop"))
		require.NoError(t, client.Delete("foo"))
		{
			expectedMailboxNames := []string{
				"INBOX",
				"foo.bar",
			}
			expectedAttributes := []string{goimap.UnmarkedAttr}
			checkMailboxesMatchNamesAndAttributes(t, client, "", "*", expectedMailboxNames, expectedAttributes)
		}
		{
			mailboxes := listMailboxesClient(t, client, "", "%")
			for _, mailbox := range mailboxes {
				if mailbox.Name == "INBOX" {
					require.ElementsMatch(t, mailbox.Attributes, []string{goimap.UnmarkedAttr})
				} else if mailbox.Name == "foo" {
					require.ElementsMatch(t, mailbox.Attributes, []string{goimap.NoSelectAttr})
				} else {
					require.Fail(t, "Unexpected mailbox name")
				}
			}
		}
		// deleting mailboxes with \Noselect Attribute is an error
		require.Error(t, client.Delete("foo"))
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

func TestDeleteMailboxCausesDisconnect(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C("b001 CREATE mbox1")
		c.S("b001 OK CREATE")

		c.C("b002 SELECT mbox1").OK("b002")
		c.C("b003 DELETE mbox1").OK("b003")
		c.S(response.Bye().WithMailboxDeleted().String())
	})
}

func TestDeleteMailboxCausesDisconnectOnOtherClients(t *testing.T) {
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
