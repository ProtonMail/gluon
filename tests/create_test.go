package tests

import (
	"testing"

	"github.com/emersion/go-imap/client"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		matchMailboxNamesClient(t, client, "", "*", []string{"INBOX"})
		require.NoError(t, client.Create("owatagusiam"))
		matchMailboxNamesClient(t, client, "", "*", []string{"INBOX", "owatagusiam"})
		require.NoError(t, client.Create("owatagusiam/blurdybloop"))
		matchMailboxNamesClient(t, client, "", "*", []string{"INBOX", "owatagusiam", "owatagusiam/blurdybloop"})
	})
}

func TestCreateEndingInSeparator(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		matchMailboxNamesClient(t, client, "", "*", []string{"INBOX"})
		require.NoError(t, client.Create("owatagusiam/"))
		matchMailboxNamesClient(t, client, "", "*", []string{"INBOX", "owatagusiam"})
	})
}

func TestCreateCannotCreateInbox(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		require.Error(t, client.Create("INBOX"))
	})
}

func TestCreateCannotCreateExistingMailbox(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		require.NoError(t, client.Create("Folder"))
		matchMailboxNamesClient(t, client, "", "*", []string{"INBOX", "Folder"})
		require.Error(t, client.Create("Folder"))
	})
}

func TestCreateWithDifferentHierarchySeparator(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		matchMailboxNamesClient(t, client, "", "*", []string{"INBOX"})
		require.NoError(t, client.Create("Folder"))
		matchMailboxNamesClient(t, client, "", "*", []string{"INBOX", "Folder"})
		require.NoError(t, client.Create("Folder\\Bar"))
		matchMailboxNamesClient(t, client, "", "*", []string{"INBOX", "Folder", "Folder\\Bar"})
	})
}

func TestCreateWithNilHierarchySeparator(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t, withDelimiter("")), func(client *client.Client, _ *testSession) {
		matchMailboxNamesClient(t, client, "", "*", []string{"INBOX"})
		require.NoError(t, client.Create("Folder/Bar"))
		matchMailboxNamesClient(t, client, "", "*", []string{"INBOX", "Folder/Bar"})
		require.NoError(t, client.Create("Folder"))
		matchMailboxNamesClient(t, client, "", "*", []string{"INBOX", "Folder", "Folder/Bar"})
	})
}

func TestCreatePreviousLevelHierarchyIfNonExisting(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		require.NoError(t, client.Create("Folder/Bar/ZZ"))
		matchMailboxNamesClient(t, client, "", "*", []string{"INBOX", "Folder", "Folder/Bar", "Folder/Bar/ZZ"})
	})
}

func TestEnsureNewMailboxWithDeletedNameHasGreaterId(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		var oldValidity uint32
		var newValidity uint32

		{
			// create Folder inbox, get id and delete
			require.NoError(t, client.Create("mbox1"))
			mailboxStatus, err := client.Select("mbox1", true)
			require.NoError(t, err)
			oldValidity = mailboxStatus.UidValidity
			require.NoError(t, client.Unselect())
			// Destroy Folder inbox
			require.NoError(t, client.Delete("mbox1"))
			require.NoError(t, client.Create("mbox2"))
		}

		{
			// re-create Folder inbox
			require.NoError(t, client.Create("mbox1"))
			mailboxStatus, err := client.Select("mbox1", true)
			require.NoError(t, err)
			newValidity = mailboxStatus.UidValidity
			require.Greater(t, newValidity, oldValidity)
			oldValidity = newValidity
		}

		{
			require.NoError(t, client.Unselect())
			require.NoError(t, client.Delete("mbox1"))
			require.NoError(t, client.Delete("mbox2"))
			require.NoError(t, client.Create("mbox2"))
			mailboxStatus, err := client.Select("mbox2", true)
			require.NoError(t, err)
			newValidity = mailboxStatus.UidValidity
			require.Greater(t, newValidity, oldValidity)
		}
	})
}

func TestCreateAdjacentSeparator(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C(`A001 create foo//bar`)
		c.Sx(`^A001 NO .*adjacent hierarchy separators\r\n$`)
	})
}

func TestCreateBeginsWithSeparator(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C(`A001 create /foo`)
		c.Sx(`^A001 NO .*begins with hierarchy separator\r\n$`)
	})
}
