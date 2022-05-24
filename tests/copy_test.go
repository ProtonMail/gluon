package tests

import (
	"testing"

	goimap "github.com/emersion/go-imap"
	uidplus "github.com/emersion/go-imap-uidplus"
	"github.com/emersion/go-imap/client"
	"github.com/stretchr/testify/require"
)

func TestCopy(t *testing.T) {
	runOneToOneTestClientWithData(t, "user", "pass", "/", func(client *client.Client, s *testSession, mbox, mboxID string) {
		{
			// There are 100 messages in the origin and no messages in the destination.
			mailboxStatus, err := client.Status(mbox, []goimap.StatusItem{goimap.StatusMessages})
			require.NoError(t, err)
			require.Equal(t, uint32(100), mailboxStatus.Messages)
		}
		uidClient := uidplus.NewClient(client)
		{
			// Copy half the messages to the destination.
			sequenceSet, seqErr := goimap.ParseSeqSet("1:50")
			require.NoError(t, seqErr)
			validity, srcUids, dstUids, err := uidClient.UidCopy(sequenceSet, "inbox")
			require.NoError(t, err)
			require.Equal(t, uint32(1), validity)
			require.Equal(t, uint32(1), srcUids.Set[0].Start)
			require.Equal(t, uint32(50), srcUids.Set[0].Stop)
			require.Equal(t, uint32(1), dstUids.Set[0].Start)
			require.Equal(t, uint32(50), dstUids.Set[0].Stop)

			// Check that 500 messages are in the new mailbox
			mailboxStatus, err := client.Status("inbox", []goimap.StatusItem{goimap.StatusMessages, goimap.StatusRecent})
			require.NoError(t, err)
			require.Equal(t, uint32(50), mailboxStatus.Messages)
			// check if recent flag was set for the copied messages
			require.Equal(t, uint32(50), mailboxStatus.Recent)
		}
		require.NoError(t, client.Noop())
		{
			// Copy the other half the messages to the destination (this time using UID COPY).
			sequenceSet, seqErr := goimap.ParseSeqSet("51:100")
			require.NoError(t, seqErr)
			validity, srcUids, dstUids, err := uidClient.UidCopy(sequenceSet, "inbox")
			require.NoError(t, err)
			require.Equal(t, uint32(1), validity)
			require.Equal(t, uint32(51), srcUids.Set[0].Start)
			require.Equal(t, uint32(100), srcUids.Set[0].Stop)
			require.Equal(t, uint32(51), dstUids.Set[0].Start)
			require.Equal(t, uint32(100), dstUids.Set[0].Stop)

			// Check that 100 messages are in the new mailbox
			mailboxStatus, err := client.Status("inbox", []goimap.StatusItem{goimap.StatusMessages, goimap.StatusRecent})
			require.NoError(t, err)
			require.Equal(t, uint32(100), mailboxStatus.Messages)
			require.Equal(t, uint32(100), mailboxStatus.Recent)

		}
	})
}

func TestCopyTryCreate(t *testing.T) {
	// Test can't be remove since there is no way to check the TRYCREATE response from the server
	runOneToOneTestWithData(t, "user", "pass", "/", func(c *testConnection, s *testSession, mbox, mboxID string) {
		// There are 100 messages in the origin.
		c.Cf(`A001 status %v (messages)`, mbox).Sxe(`MESSAGES 100`).OK(`A001`)

		// Copy to a nonexistent destination.
		c.C(`A002 copy 1:* this-name-does-not-exist`)
		c.Sx(`A002 NO \[TRYCREATE\]`)

		// UID COPY to a nonexistent destination.
		c.C(`A002 uid copy 1:* this-name-does-not-exist`)
		c.Sx(`A002 NO \[TRYCREATE\]`)
	})
}

func TestCopyNonExistingClient(t *testing.T) {
	runOneToOneTestClientWithData(t, "user", "pass", "/", func(client *client.Client, s *testSession, mbox, mboxID string) {
		{
			// Move message intervals to inbox
			sequenceSet, seqErr := goimap.ParseSeqSet("1:25,76:100")
			require.NoError(t, seqErr)
			require.NoError(t, client.Move(sequenceSet, "inbox"))

			// Check that 50 messages are in the new mailbox
			mailboxStatus, err := client.Status("inbox", []goimap.StatusItem{goimap.StatusMessages})
			require.NoError(t, err)
			require.Equal(t, uint32(50), mailboxStatus.Messages)
		}
		{
			// Attempting to UID COPY nonexistent messages with UIDs lower than the smallest in the mailbox returns OK.
			sequenceSet, seqErr := goimap.ParseSeqSet("51:100")
			require.NoError(t, seqErr)
			require.Error(t, client.Copy(sequenceSet, "inbox"))
		}
		{
			// Attempting to UID COPY nonexistent messages with UIDs lower than the smallest in the mailbox returns OK.
			uidClient := uidplus.NewClient(client)
			sequenceSet, seqErr := goimap.ParseSeqSet("1:25")
			require.NoError(t, seqErr)
			_, _, _, err := uidClient.UidCopy(sequenceSet, "inbox")
			require.NoError(t, err)
		}
		{
			// Attempting to UID COPY nonexistent messages with UIDs lower than the smallest in the mailbox returns OK.
			uidClient := uidplus.NewClient(client)
			sequenceSet, seqErr := goimap.ParseSeqSet("76:100")
			require.NoError(t, seqErr)
			_, _, _, err := uidClient.UidCopy(sequenceSet, "inbox")
			require.NoError(t, err)
		}
	})
}

func TestCopyNonExisting(t *testing.T) {
	runOneToOneTestWithData(t, "user", "pass", "/", func(c *testConnection, s *testSession, mbox, mboxID string) {
		// MOVE some of the messages out of the mailbox.
		c.C(`A001 move 1:24,76:100 inbox`).OK(`A001`)

		// Attempting to MOVE nonexistent messages by sequence number returns BAD.
		c.C(`A002 move 51:100 inbox`).BAD(`A002`)

		// Attempting to UID MOVE nonexistent messages with UIDs lower than the smallest in the mailbox returns OK.
		c.C(`A003 uid move 1:24 inbox`)
		// Nothing should be returned
		c.Sx(`A003 OK .*`)

		// Attempting to UID MOVE nonexistent messages with UIDs higher than the largest in the mailbox returns OK.
		c.C(`A004 uid copy 76:100 inbox`)
		// Nothing should be returned
		c.Sx(`A004 OK .*`)

		c.C(`A005 uid copy 24:26 inbox`).OK(`A005`)
	})
}
