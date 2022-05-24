package tests

import (
	"testing"

	"github.com/emersion/go-imap/client"
	"github.com/stretchr/testify/require"
)

func TestRename(t *testing.T) {
	runOneToOneTestClientWithAuth(t, "user", "pass", "/", func(client *client.Client, _ *testSession) {
		require.NoError(t, client.Create("blurdybloop"))
		require.NoError(t, client.Create("foo"))
		require.NoError(t, client.Create("foo/bar"))

		matchMailboxNamesClient(t, client, "", "*", []string{"INBOX", "foo", "foo/bar", "blurdybloop"})

		require.NoError(t, client.Rename("blurdybloop", "sarasoop"))
		require.NoError(t, client.Rename("foo", "zowie"))

		matchMailboxNamesClient(t, client, "", "*", []string{"INBOX", "zowie", "zowie/bar", "sarasoop"})
	})
}

func TestRenameHierarchy(t *testing.T) {
	runOneToOneTestClientWithAuth(t, "user", "pass", "/", func(client *client.Client, _ *testSession) {
		require.NoError(t, client.Create("foo/bar/zap"))

		matchMailboxNamesClient(t, client, "", "*", []string{"INBOX", "foo", "foo/bar", "foo/bar/zap"})

		require.NoError(t, client.Rename("foo/bar/zap", "baz/rag/zowie"))

		matchMailboxNamesClient(t, client, "", "*", []string{"INBOX", "foo", "foo/bar", "baz", "baz/rag", "baz/rag/zowie"})
	})
}

func TestRenameInbox(t *testing.T) {
	runOneToOneTestWithData(t, "user", "pass", "/", func(c *testConnection, s *testSession, mbox, mboxID string) {
		// Put all the 100 messages into the inbox.
		c.C("tag move 1:* inbox").OK("tag")
		c.C("tag status inbox (messages)").Sxe("MESSAGES 100").OK("tag")

		// Renaming the inbox will create a new mailbox and put all the messages in there.
		c.C("tag rename inbox some/other/mailbox").OK("tag")

		// There are now 0 messages in the inbox; they've all been moved to this new mailbox.
		c.C("tag status inbox (messages)").Sxe("MESSAGES 0").OK("tag")
		c.C("tag status some/other/mailbox (messages)").Sxe("MESSAGES 100").OK("tag")

		// Put all the 100 messages back into the inbox.
		c.C("tag select some/other/mailbox").OK("tag")
		c.C("tag move 1:* inbox").OK("tag")
		c.C("tag status inbox (messages)").Sxe("MESSAGES 100").OK("tag")
		c.C("tag status some/other/mailbox (messages)").Sxe("MESSAGES 0").OK("tag")

		// Renaming the inbox again will do the same as before.
		c.C("tag rename inbox yet/another/mailbox").OK("tag")
		c.C("tag status inbox (messages)").Sxe("MESSAGES 0").OK("tag")
		c.C("tag status some/other/mailbox (messages)").Sxe("MESSAGES 0").OK("tag")
		c.C("tag status yet/another/mailbox (messages)").Sxe("MESSAGES 100").OK("tag")
	})
}
