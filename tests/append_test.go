package tests

import (
	"testing"
	"time"

	goimap "github.com/emersion/go-imap"
	uidplus "github.com/emersion/go-imap-uidplus"
	"github.com/emersion/go-imap/client"
	"github.com/stretchr/testify/require"
)

func TestAppend(t *testing.T) {
	const (
		mailboxName = "saved-messages"
		messagePath = "testdata/afternoon-meeting.eml"
	)

	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		require.NoError(t, client.Create(mailboxName))
		{
			// first check, empty mailbox
			status, err := client.Status(mailboxName, []goimap.StatusItem{goimap.StatusMessages})
			require.NoError(t, err)
			require.Equal(t, uint32(0), status.Messages, "Expected message count does not match")
		}
		{
			require.NoError(t, doAppendWithClientFromFile(t, client, mailboxName, messagePath, time.Now(), goimap.SeenFlag))
			require.NoError(t, doAppendWithClientFromFile(t, client, mailboxName, messagePath, time.Now(), goimap.SeenFlag))
			require.NoError(t, doAppendWithClientFromFile(t, client, mailboxName, messagePath, time.Now(), goimap.SeenFlag))
			require.Error(t, doAppendWithClientFromFile(t, client, mailboxName, messagePath, time.Now(), goimap.RecentFlag))
		}
		{
			// second check, there should be 3 messages
			status, err := client.Status(mailboxName, []goimap.StatusItem{goimap.StatusMessages})
			require.NoError(t, err)
			require.Equal(t, uint32(3), status.Messages, "Expected message count does not match")
		}
	})
}

func TestAppendWithUidPlus(t *testing.T) {
	const (
		mailboxName = "saved-messages"
		messagePath = "testdata/afternoon-meeting.eml"
	)

	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		const (
			validityUid = uint32(1)
		)
		require.NoError(t, client.Create(mailboxName))
		{
			// first check, empty mailbox
			status, err := client.Status(mailboxName, []goimap.StatusItem{goimap.StatusMessages})
			require.NoError(t, err)
			require.Equal(t, uint32(0), status.Messages, "Expected message count does not match")
		}
		{
			// Insert some messages
			clientUidPlus := uidplus.NewClient(client)
			{
				validity, uid, err := doAppendWithClientPlusFromFile(t, clientUidPlus, mailboxName, messagePath, goimap.SeenFlag)
				require.NoError(t, err)
				require.Equal(t, validityUid, validity)
				require.Equal(t, uint32(1), uid)
			}
			{
				validity, uid, err := doAppendWithClientPlusFromFile(t, clientUidPlus, mailboxName, messagePath, goimap.SeenFlag)
				require.NoError(t, err)
				require.Equal(t, validityUid, validity)
				require.Equal(t, uint32(2), uid)
			}
			{
				validity, uid, err := doAppendWithClientPlusFromFile(t, clientUidPlus, mailboxName, messagePath, goimap.SeenFlag)
				require.NoError(t, err)
				require.Equal(t, validityUid, validity)
				require.Equal(t, uint32(3), uid)
			}
			require.Error(t, doAppendWithClientFromFile(t, client, mailboxName, messagePath, time.Now(), goimap.RecentFlag))
		}
		{
			// second check, there should be 3 messages
			status, err := client.Status(mailboxName, []goimap.StatusItem{goimap.StatusMessages})
			require.NoError(t, err)
			require.Equal(t, uint32(3), status.Messages, "Expected message count does not match")
		}
	})
}

func TestAppendNoSuchMailbox(t *testing.T) {
	const (
		mailboxName = "saved-messages"
		messagePath = "testdata/afternoon-meeting.eml"
	)

	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		err := doAppendWithClientFromFile(t, client, mailboxName, messagePath, time.Now(), goimap.SeenFlag)
		require.Error(t, err)
	})
	// Run the old test as well as there is no way to check for the `[TRYCREATE]` flag of the response
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.doAppendFromFile(`saved-messages`, `testdata/afternoon-meeting.eml`, `\Seen`).expect(`NO \[TRYCREATE\]`)
	})
}

func TestAppendWhileSelected(t *testing.T) {
	const (
		mailboxName = "saved-messages"
		messagePath = "testdata/afternoon-meeting.eml"
	)

	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		require.NoError(t, client.Create(mailboxName))
		// Mailbox should have read-write modifier
		mailboxStatus, err := client.Select(mailboxName, false)
		require.NoError(t, err)
		require.Equal(t, false, mailboxStatus.ReadOnly)
		// Add new message
		require.NoError(t, doAppendWithClientFromFile(t, client, mailboxName, messagePath, time.Now(), goimap.SeenFlag))
		// EXISTS response is assigned to Messages member
		require.Equal(t, uint32(1), client.Mailbox().Messages)
		// Check if added
		status, err := client.Status(mailboxName, []goimap.StatusItem{goimap.StatusMessages})
		require.NoError(t, err)
		require.Equal(t, uint32(1), status.Messages, "Expected message count does not match")
	})
}
