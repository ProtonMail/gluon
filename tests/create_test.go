package tests

import (
	"fmt"
	"github.com/ProtonMail/gluon/imap"
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

func TestCreate_UIDValidity_Bumped(t *testing.T) {
	uidValidityGenerator := imap.NewIncrementalUIDValidityGenerator()
	runServer(t, defaultServerOptions(t, withUIDValidityGenerator(uidValidityGenerator)), func(s *testSession) {

		currentUIDValidity := uidValidityGenerator.GetValue()

		// Create some mailboxes; they'll have the initial UID validity of 1.
		s.withConnection(s.options.defaultUsername(), func(c *testConnection) {
			c.C("tag create a").OK("tag")
			c.C("tag create b").OK("tag")
			c.C("tag create c").OK("tag")

			currentUIDValidity++
			c.C("tag select a").Sxe(fmt.Sprintf("UIDVALIDITY %v", currentUIDValidity)).OK("tag")
			currentUIDValidity++
			c.C("tag select b").Sxe(fmt.Sprintf("UIDVALIDITY %v", currentUIDValidity)).OK("tag")
			currentUIDValidity++
			c.C("tag select c").Sxe(fmt.Sprintf("UIDVALIDITY %v", currentUIDValidity)).OK("tag")
		})

		// Delete the mailboxes.
		s.withConnection(s.options.defaultUsername(), func(c *testConnection) { c.C("tag delete a").OK("tag") })
		s.withConnection(s.options.defaultUsername(), func(c *testConnection) { c.C("tag delete b").OK("tag") })
		s.withConnection(s.options.defaultUsername(), func(c *testConnection) { c.C("tag delete c").OK("tag") })

		// Recreate the mailboxes; they'll have a new UID validity.
		s.withConnection(s.options.defaultUsername(), func(c *testConnection) {
			c.C("tag create a").OK("tag")
			c.C("tag create b").OK("tag")
			c.C("tag create c").OK("tag")

			currentUIDValidity++
			c.C("tag select a").Sxe(fmt.Sprintf("UIDVALIDITY %v", currentUIDValidity)).OK("tag")
			currentUIDValidity++
			c.C("tag select b").Sxe(fmt.Sprintf("UIDVALIDITY %v", currentUIDValidity)).OK("tag")
			currentUIDValidity++
			c.C("tag select c").Sxe(fmt.Sprintf("UIDVALIDITY %v", currentUIDValidity)).OK("tag")
		})

		// Bump the global UID validity.
		s.uidValidityBumped(s.options.defaultUsername())

		// Ensure the UID validity has been bumped.
		s.flush(s.options.defaultUsername())

		// The mailboxes should all have a new UID validity.
		s.withConnection(s.options.defaultUsername(), func(c *testConnection) {
			currentUIDValidity = 14
			c.C("tag select a").Sxe(fmt.Sprintf("UIDVALIDITY %v", currentUIDValidity)).OK("tag")
			currentUIDValidity++
			c.C("tag select b").Sxe(fmt.Sprintf("UIDVALIDITY %v", currentUIDValidity)).OK("tag")
			currentUIDValidity++
			c.C("tag select c").Sxe(fmt.Sprintf("UIDVALIDITY %v", currentUIDValidity)).OK("tag")
		})
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
