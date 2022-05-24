package tests

import (
	"testing"

	"github.com/emersion/go-imap/client"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	runOneToOneTestClientWithAuth(t, "user", "pass", "/", func(client *client.Client, _ *testSession) {
		matchMailboxNamesClient(t, client, "", "*", []string{"INBOX"})
		require.NoError(t, client.Create("owatagusiam"))
		matchMailboxNamesClient(t, client, "", "*", []string{"INBOX", "owatagusiam"})
		require.NoError(t, client.Create("owatagusiam/blurdybloop"))
		matchMailboxNamesClient(t, client, "", "*", []string{"INBOX", "owatagusiam", "owatagusiam/blurdybloop"})
	})
}

func TestCreateEndingInSeparator(t *testing.T) {
	runOneToOneTestClientWithAuth(t, "user", "pass", "/", func(client *client.Client, _ *testSession) {
		matchMailboxNamesClient(t, client, "", "*", []string{"INBOX"})
		require.NoError(t, client.Create("owatagusiam/"))
		matchMailboxNamesClient(t, client, "", "*", []string{"INBOX", "owatagusiam"})
	})
}

func TestCreateCannotCreateInbox(t *testing.T) {
	runOneToOneTestClientWithAuth(t, "user", "pass", "/", func(client *client.Client, _ *testSession) {
		require.Error(t, client.Create("INBOX"))
	})
}

func TestCreateCannotCreateExistingMailbox(t *testing.T) {
	runOneToOneTestClientWithAuth(t, "user", "pass", "/", func(client *client.Client, _ *testSession) {
		require.NoError(t, client.Create("Folder"))
		matchMailboxNamesClient(t, client, "", "*", []string{"INBOX", "Folder"})
		require.Error(t, client.Create("Folder"))
	})
}

func TestCreateWithDifferentHierarchySeparator(t *testing.T) {
	runOneToOneTestClientWithAuth(t, "user", "pass", "/", func(client *client.Client, _ *testSession) {
		matchMailboxNamesClient(t, client, "", "*", []string{"INBOX"})
		require.NoError(t, client.Create("Folder"))
		matchMailboxNamesClient(t, client, "", "*", []string{"INBOX", "Folder"})
		require.NoError(t, client.Create("Folder\\Bar"))
		matchMailboxNamesClient(t, client, "", "*", []string{"INBOX", "Folder", "Folder\\Bar"})
	})
}

func TestCreatePreviousLevelHierarchyIfNonExisting(t *testing.T) {
	runOneToOneTestClientWithAuth(t, "user", "pass", "/", func(client *client.Client, _ *testSession) {
		require.NoError(t, client.Create("Folder/Bar/ZZ"))
		matchMailboxNamesClient(t, client, "", "*", []string{"INBOX", "Folder", "Folder/Bar", "Folder/Bar/ZZ"})
	})
}

// TODO: GOMSRV-51.
func _TestEnsureNewMailboxWithDeletedNameHasGreaterId(t *testing.T) {
	runOneToOneTestClientWithAuth(t, "user", "pass", "/", func(client *client.Client, _ *testSession) {
		const (
			inboxName = "Folder"
		)
		var firstId uint32
		var secondId uint32
		{
			// create Folder inbox, get id and delete
			require.NoError(t, client.Create(inboxName))
			mailboxStatus, err := client.Select(inboxName, true)
			require.NoError(t, err)
			firstId = mailboxStatus.UidValidity
			require.NoError(t, client.Unselect())
			// Destroy Folder inbox
			require.NoError(t, client.Delete(inboxName))
		}
		{
			// re-create Folder inbox
			require.NoError(t, client.Create(inboxName))
			mailboxStatus, err := client.Select(inboxName, true)
			require.NoError(t, err)
			secondId = mailboxStatus.UidValidity
		}
		require.NotEqual(t, secondId, firstId)
	})
}
