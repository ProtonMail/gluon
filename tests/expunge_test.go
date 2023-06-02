package tests

import (
	"fmt"
	"strings"
	"testing"
	"time"

	goimap "github.com/emersion/go-imap"
	uidplus "github.com/emersion/go-imap-uidplus"
	"github.com/emersion/go-imap/client"
	"github.com/stretchr/testify/require"
)

func TestExpungeUnmarked(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		require.NoError(t, client.Create("mbox"))
		_, err := client.Select("mbox", false)
		require.NoError(t, err)
		require.NoError(t, client.Append("mbox", []string{goimap.SeenFlag}, time.Now(), strings.NewReader(buildRFC5322TestLiteral("To: 1@pm.me"))))
	})
}

func TestExpungeSingle(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		require.NoError(t, client.Create("mbox"))
		_, err := client.Select("mbox", false)
		require.NoError(t, err)

		require.NoError(t, client.Append("mbox", []string{goimap.SeenFlag}, time.Now(), strings.NewReader(buildRFC5322TestLiteral("To: 1@pm.me"))))

		messages := storeWithRetrievalClient(t, client, createSeqSet("1"), goimap.AddFlags, []interface{}{goimap.DeletedFlag})
		require.Equal(t, 1, len(messages))
		require.ElementsMatch(t, messages[0].Flags, []string{goimap.SeenFlag, goimap.DeletedFlag, goimap.RecentFlag})
		expungedIds := expungeClient(t, client)
		require.Equal(t, 1, len(expungedIds))
		require.Equal(t, uint32(1), expungedIds[0])
	})
}

func TestRemoveSameMessageTwice(t *testing.T) {
	runManyToOneTestWithAuth(t, defaultServerOptions(t), []int{1, 2, 3}, func(c map[int]*testConnection, s *testSession) {
		// There are three messages in the inbox.
		c[1].doAppend(`inbox`, buildRFC5322TestLiteral(`To: 1@pm.me`)).expect("OK")
		c[1].doAppend(`inbox`, buildRFC5322TestLiteral(`To: 2@pm.me`)).expect("OK")
		c[1].doAppend(`inbox`, buildRFC5322TestLiteral(`To: 3@pm.me`)).expect("OK")

		// There are three clients selected in the inbox.
		for i := range c {
			c[i].C("tag select inbox").OK("tag")
		}

		// First and second client both mark the first message as deleted.
		c[1].C(`tag store 1 +flags (\deleted)`).OK("tag")
		c[2].C(`tag store 1 +flags (\deleted)`).OK("tag")

		// Ignore the fetch responses of the third client.
		c[3].C("tag noop").OK("tag")

		// First and second client both expunge the first message.
		c[1].C(`tag expunge`).OK("tag")
		c[2].C(`tag expunge`).OK("tag")

		// Third client should just get one expunge response.
		c[3].C("tag noop")
		c[3].S(`* 1 EXPUNGE`)
		c[3].OK("tag")
	})
}

func TestExpungeInterval(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		require.NoError(t, client.Create("mbox"))
		_, err := client.Select("mbox", false)
		require.NoError(t, err)
		for i := 1; i <= 4; i++ {
			require.NoError(t, client.Append("mbox", []string{goimap.SeenFlag}, time.Now(), strings.NewReader(buildRFC5322TestLiteral(fmt.Sprintf(`To: %d@pm.me`, i)))))
		}
		messages := storeWithRetrievalClient(t, client, createSeqSet("1,3"), goimap.AddFlags, []interface{}{goimap.DeletedFlag})
		require.Equal(t, 2, len(messages))
		for _, message := range messages {
			require.ElementsMatch(t, message.Flags, []string{goimap.SeenFlag, goimap.DeletedFlag, goimap.RecentFlag})
		}
		expungedIds := expungeClient(t, client)
		require.Equal(t, 2, len(expungedIds))
		require.Equal(t, expungedIds, []uint32{1, 2})
	})
}

func beforeOrAfterExpungeCheck(t *testing.T, client *client.Client, mailboxName string) {
	// Shared code used to for checking the mailbox state after expunge for the
	// TestExpungeWithAppendBeforeMailboxSelect and TestExpungeWithAppendAfterMailboxSelect
	{
		messages := storeWithRetrievalClient(t, client, createSeqSet("1"), goimap.AddFlags, []interface{}{goimap.DeletedFlag})
		require.Equal(t, 1, len(messages))
		require.ElementsMatch(t, messages[0].Flags, []string{goimap.SeenFlag, goimap.RecentFlag, goimap.DeletedFlag})
	}
	{
		messages := storeWithRetrievalClient(t, client, createSeqSet("2"), goimap.AddFlags, []interface{}{goimap.DeletedFlag})
		require.Equal(t, 1, len(messages))
		require.ElementsMatch(t, messages[0].Flags, []string{goimap.RecentFlag, goimap.DeletedFlag})
	}
	{
		messages := storeWithRetrievalClient(t, client, createSeqSet("3"), goimap.AddFlags, []interface{}{goimap.DeletedFlag})
		require.Equal(t, 1, len(messages))
		require.ElementsMatch(t, messages[0].Flags, []string{goimap.SeenFlag, goimap.RecentFlag, goimap.DeletedFlag})
	}

	{
		expungedIds := expungeClient(t, client)
		require.Equal(t, []uint32{1, 1, 1}, expungedIds)
	}

	// There are 2 messages in saved-messages.
	mailboxStatus, err := client.Status(mailboxName, []goimap.StatusItem{goimap.StatusMessages})
	require.NoError(t, err)
	require.Equal(t, uint32(2), mailboxStatus.Messages)
}

func TestExpungeWithAppendBeforeMailboxSelect(t *testing.T) {
	// This test exists to catch an issue where in the past messages that might be added before/after selected would
	// not be removed.
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		const mailboxName = "saved-messages"
		require.NoError(t, client.Create(mailboxName))

		require.NoError(t, client.Append(mailboxName, []string{goimap.SeenFlag}, time.Now(), strings.NewReader(buildRFC5322TestLiteral("To: 1@pm.me"))))
		require.NoError(t, client.Append(mailboxName, []string{}, time.Now(), strings.NewReader(buildRFC5322TestLiteral("To: 2@pm.me"))))
		require.NoError(t, client.Append(mailboxName, []string{goimap.SeenFlag}, time.Now(), strings.NewReader(buildRFC5322TestLiteral("To: 3@pm.me"))))
		require.NoError(t, client.Append(mailboxName, []string{}, time.Now(), strings.NewReader(buildRFC5322TestLiteral("To: 4@pm.me"))))
		require.NoError(t, client.Append(mailboxName, []string{goimap.SeenFlag}, time.Now(), strings.NewReader(buildRFC5322TestLiteral("To: 5@pm.me"))))

		_, err := client.Select(mailboxName, false)
		require.NoError(t, err)

		beforeOrAfterExpungeCheck(t, client, mailboxName)
	})
}

func TestExpungeWithAppendAfterMailBoxSelect(t *testing.T) {
	// This test exists to catch an issue where in the past messages that might be added before/after selected would
	// not be removed.
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		const mailboxName = "saved-messages"

		require.NoError(t, client.Create(mailboxName))
		_, err := client.Select(mailboxName, false)
		require.NoError(t, err)

		require.NoError(t, client.Append(mailboxName, []string{goimap.SeenFlag}, time.Now(), strings.NewReader(buildRFC5322TestLiteral("To: 1@pm.me"))))
		require.NoError(t, client.Append(mailboxName, []string{}, time.Now(), strings.NewReader(buildRFC5322TestLiteral("To: 2@pm.me"))))
		require.NoError(t, client.Append(mailboxName, []string{goimap.SeenFlag}, time.Now(), strings.NewReader(buildRFC5322TestLiteral("To: 3@pm.me"))))
		require.NoError(t, client.Append(mailboxName, []string{}, time.Now(), strings.NewReader(buildRFC5322TestLiteral("To: 4@pm.me"))))
		require.NoError(t, client.Append(mailboxName, []string{goimap.SeenFlag}, time.Now(), strings.NewReader(buildRFC5322TestLiteral("To: 5@pm.me"))))

		beforeOrAfterExpungeCheck(t, client, mailboxName)
	})
}

func TestExpungeUID(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		const mailboxName = "mbox"

		require.NoError(t, client.Create(mailboxName))
		_, err := client.Select(mailboxName, false)
		require.NoError(t, err)

		require.NoError(t, client.Append(mailboxName, []string{goimap.SeenFlag}, time.Now(), strings.NewReader(buildRFC5322TestLiteral("To: 1@pm.me"))))
		uidClient := uidplus.NewClient(client)

		{
			messages := storeWithRetrievalClient(t, client, createSeqSet("1"), goimap.AddFlags, []interface{}{goimap.DeletedFlag})
			require.Equal(t, 1, len(messages))
			require.ElementsMatch(t, messages[0].Flags, []string{goimap.SeenFlag, goimap.RecentFlag, goimap.DeletedFlag})
		}
		{
			messages := uidFetchMessagesClient(t, client, createSeqSet("1"), []goimap.FetchItem{goimap.FetchUid, goimap.FetchFlags})
			require.Equal(t, 1, len(messages))
			require.ElementsMatch(t, messages[0].Flags, []string{goimap.SeenFlag, goimap.RecentFlag, goimap.DeletedFlag})
		}

		{
			expungedIds := uidExpungeClient(t, uidClient, createSeqSet("1"))
			require.Equal(t, 1, len(expungedIds))
			require.Equal(t, uint32(1), expungedIds[0])
		}

		{
			mailboxStatus, err := client.Status(mailboxName, []goimap.StatusItem{goimap.StatusMessages})
			require.NoError(t, err)
			require.Equal(t, uint32(0), mailboxStatus.Messages)
		}

		// Append new messages
		for i := 1; i <= 4; i++ {
			require.NoError(t, client.Append(mailboxName, []string{goimap.SeenFlag}, time.Now(), strings.NewReader(buildRFC5322TestLiteral(fmt.Sprintf(`To: %d@pm.me`, i)))))
		}

		{
			messages := uidStoreWithRetrievalClient(t, client, createSeqSet("2,4"), goimap.AddFlags, []interface{}{goimap.DeletedFlag})
			require.Equal(t, 2, len(messages))
			for _, message := range messages {
				require.ElementsMatch(t, message.Flags, []string{goimap.SeenFlag, goimap.RecentFlag, goimap.DeletedFlag})
			}
		}
		{
			expungedIds := uidExpungeClient(t, uidClient, createSeqSet("4,2"))
			require.Equal(t, 2, len(expungedIds))
			require.Equal(t, []uint32{1, 2}, expungedIds)
		}

		{
			mailboxStatus, err := client.Status(mailboxName, []goimap.StatusItem{goimap.StatusMessages})
			require.NoError(t, err)
			require.Equal(t, uint32(2), mailboxStatus.Messages)
			require.Equal(t, 2, len(fetchMessagesClient(t, client, createSeqSet("1,2"), []goimap.FetchItem{goimap.FetchUid})))
		}
	})
}

func TestExpungeResponseSequence(t *testing.T) {
	// Test the response sequence produce by the expunge command, should match the example output of
	// rfc3501#section-6.4.3
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		const mailboxName = "mbox"
		require.NoError(t, client.Create(mailboxName))
		_, err := client.Select(mailboxName, false)
		require.NoError(t, err)

		for i := 1; i <= 11; i++ {
			require.NoError(t, client.Append(mailboxName, []string{goimap.SeenFlag}, time.Now(), strings.NewReader(buildRFC5322TestLiteral(fmt.Sprintf(`To: %d@pm.me`, i)))))
		}

		{
			messages := storeWithRetrievalClient(t, client, createSeqSet("3,4,7,11"), goimap.AddFlags, []interface{}{goimap.DeletedFlag})
			require.Equal(t, 4, len(messages))
			for _, message := range messages {
				require.ElementsMatch(t, message.Flags, []string{goimap.SeenFlag, goimap.RecentFlag, goimap.DeletedFlag})
			}
		}
		{
			expungedIds := expungeClient(t, client)
			require.Equal(t, 4, len(expungedIds))
			require.Equal(t, []uint32{3, 3, 5, 8}, expungedIds)
		}
	})
}
