package tests

import (
	"testing"

	goimap "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/stretchr/testify/require"
)

func checkMailboxesMatchNamesAndAttributes(t *testing.T, client *client.Client, reference string, expression string, expectedNames []string, expectedAttributes []string) {
	mailboxes := listMailboxesClient(t, client, "", "*")

	var actualMailboxNames []string

	for _, mailbox := range mailboxes {
		actualMailboxNames = append(actualMailboxNames, mailbox.Name)
		require.ElementsMatch(t, mailbox.Attributes, expectedAttributes)
	}

	require.ElementsMatch(t, actualMailboxNames, expectedNames)
}

func TestDelete(t *testing.T) {
	runOneToOneTestClientWithAuth(t, "user", "pass", ".", func(client *client.Client, _ *testSession) {
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
					// Case conflict , goimap.NoSelectAttr is "Noselect"
					require.ElementsMatch(t, mailbox.Attributes, []string{"\\NoSelect"})
				} else {
					require.Fail(t, "Unexpected mailbox name")
				}
			}
		}
		// deleting mailboxes with \NoSelect Attribute is an error
		require.Error(t, client.Delete("foo"))
	})
}

func TestDeleteCannotDeleteInbox(t *testing.T) {
	runOneToOneTestClientWithAuth(t, "user", "pass", "/", func(client *client.Client, _ *testSession) {
		require.Error(t, client.Delete("INBOX"))
	})
}

func TestDeleteCannotDeleteMissingMailbox(t *testing.T) {
	runOneToOneTestClientWithAuth(t, "user", "pass", "/", func(client *client.Client, _ *testSession) {
		require.Error(t, client.Delete("this doesn't exist"))
	})
}
