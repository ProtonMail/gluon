package tests

import (
	"testing"
	"time"

	imap2 "github.com/ProtonMail/gluon/imap"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestRename(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
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
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		require.NoError(t, client.Create("foo/bar/zap"))

		matchMailboxNamesClient(t, client, "", "*", []string{"INBOX", "foo", "foo/bar", "foo/bar/zap"})

		require.NoError(t, client.Rename("foo/bar/zap", "baz/rag/zowie"))

		matchMailboxNamesClient(t, client, "", "*", []string{"INBOX", "foo", "foo/bar", "baz", "baz/rag", "baz/rag/zowie"})
	})
}

func TestRenameAddHierarchy(t *testing.T) {
	type renameTC struct {
		src       string
		dest      string
		result    []string
		newFolder string
	}

	testCases := []renameTC{
		// 0 - rename the first level.
		{"foo", "bar.foo",
			[]string{"INBOX", "bar", "bar.foo", "bar.foo.bar"}, "bar"},
		// 1 - rename the last level.
		{"foo.bar", "foo.rag.bar",
			[]string{"INBOX", "foo", "foo.rag", "foo.rag.bar"}, "foo.rag"},
	}
	initialMailbox := []string{"INBOX", "foo", "foo.bar"}

	const messagePath = "testdata/afternoon-meeting.eml"

	for i, tc := range testCases {
		logrus.Trace(" --- test case ", i, " ---")
		runOneToOneTestClientWithAuth(t, defaultServerOptions(t, withDelimiter(".")), func(client *client.Client, _ *testSession) {
			require.NoError(t, client.Create("foo.bar"))
			matchMailboxNamesClient(t, client, "", "*", initialMailbox)

			// add a mail to every existing mailbox.
			for _, box := range initialMailbox {
				require.NoError(t, doAppendWithClientFromFile(t, client, box, messagePath, time.Now()))
			}

			// rename.
			require.NoError(t, client.Rename(tc.src, tc.dest))
			matchMailboxNamesClient(t, client, "", "*", tc.result)

			// all box except new one should have a mail
			for _, name := range tc.result {
				currStatus, err := client.Status(name, []imap.StatusItem{imap.StatusMessages})
				require.NoError(t, err)
				if name == tc.newFolder {
					require.Equal(t, uint32(0), currStatus.Messages, "Expected no message in the new folder %v", name)
				} else {
					require.Equal(t, uint32(1), currStatus.Messages, "Expected message to be kept in %v", name)
				}
			}
		})
	}
}

func TestRenameBadHierarchy(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t, withDelimiter(".")), func(client *client.Client, _ *testSession) {
		require.NoError(t, client.Create("foo.bar"))
		matchMailboxNamesClient(t, client, "", "*", []string{"INBOX", "foo", "foo.bar"})
		require.Error(t, client.Rename("foo", "foo.foo"))
		require.Error(t, client.Rename("foo", "foo.foo.foo"))
		require.Error(t, client.Rename("foo", "foo.bar"))
		require.NoError(t, client.Rename("foo", "bar.foo"))
	})
}

func TestRenameInbox(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap2.LabelID) {
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

	runOneToOneTestClientWithAuth(t, defaultServerOptions(t, withDelimiter(".")), func(client *client.Client, _ *testSession) {
		require.NoError(t, client.Create("INBOX.foo.bar"))
		matchMailboxNamesClient(t, client, "", "*", []string{"INBOX", "INBOX.foo", "INBOX.foo.bar"})

		// rename.
		require.NoError(t, client.Rename("INBOX", "bar"))
		matchMailboxNamesClient(t, client, "", "*", []string{"INBOX", "bar", "INBOX.foo", "INBOX.foo.bar"})
	})
}
