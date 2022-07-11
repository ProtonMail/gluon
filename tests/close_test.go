package tests

import (
	"strings"
	"testing"
	"time"

	goimap "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/stretchr/testify/require"
)

func TestClose(t *testing.T) {
	// This test is still useful as we have no way of checking the fetch responses after the store command.
	// Additionally, we also need to ensure that there are no unilateral EXPUNGE messages returned from the server after close.
	// There is currently no way to check for this with the go imap client.
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C("b001 CREATE saved-messages")
		c.S("b001 OK CREATE")

		c.doAppend(`saved-messages`, `To: 1@pm.me`, `\Seen`).expect("OK")
		c.doAppend(`saved-messages`, `To: 2@pm.me`).expect("OK")
		c.doAppend(`saved-messages`, `To: 3@pm.me`, `\Seen`).expect("OK")
		c.doAppend(`saved-messages`, `To: 4@pm.me`).expect("OK")
		c.doAppend(`saved-messages`, `To: 5@pm.me`, `\Seen`).expect("OK")

		c.C(`A002 SELECT saved-messages`)
		c.Se(`A002 OK [READ-WRITE] SELECT`)

		// TODO: Match flags in any order.
		c.C(`A003 STORE 1 +FLAGS (\Deleted)`)
		c.S(`* 1 FETCH (FLAGS (\Deleted \Recent \Seen))`)
		c.Sx("A003 OK.*")

		c.C(`A004 STORE 2 +FLAGS (\Deleted)`)
		c.S(`* 2 FETCH (FLAGS (\Deleted \Recent))`)
		c.Sx("A004 OK.*")

		c.C(`A005 STORE 3 +FLAGS (\Deleted)`)
		c.S(`* 3 FETCH (FLAGS (\Deleted \Recent \Seen))`)
		c.Sx("A005 OK.*")

		// TODO: GOMSRV-106 - Ensure this also works for cases where multiple clients have the same mailbox open
		c.C(`A202 CLOSE`)
		c.S("A202 OK CLOSE")

		// There are 2 messages in saved-messages.
		c.C(`A006 STATUS saved-messages (MESSAGES)`)
		c.S(`* STATUS "saved-messages" (MESSAGES 2)`)
		c.S(`A006 OK STATUS`)
	})
}

func TestCloseWithClient(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		const (
			messageBoxName = "saved-messages"
		)
		require.NoError(t, client.Create(messageBoxName))

		require.NoError(t, client.Append(messageBoxName, []string{goimap.SeenFlag}, time.Now(), strings.NewReader("To: 1@pm.me")))
		require.NoError(t, client.Append(messageBoxName, nil, time.Now(), strings.NewReader("To: 2@pm.me")))
		require.NoError(t, client.Append(messageBoxName, []string{goimap.SeenFlag}, time.Now(), strings.NewReader("To: 3@pm.me")))
		require.NoError(t, client.Append(messageBoxName, nil, time.Now(), strings.NewReader("To: 4@pm.me")))
		require.NoError(t, client.Append(messageBoxName, []string{goimap.SeenFlag}, time.Now(), strings.NewReader("To: 5@pm.me")))

		{
			mailboxStatus, err := client.Select(messageBoxName, false)
			require.NoError(t, err)
			require.Equal(t, false, mailboxStatus.ReadOnly)
		}

		{
			sequenceSet, _ := goimap.ParseSeqSet("1")
			require.NoError(t, client.Store(sequenceSet, goimap.AddFlags, []interface{}{goimap.DeletedFlag}, nil))
		}
		{
			sequenceSet, _ := goimap.ParseSeqSet("2")
			require.NoError(t, client.Store(sequenceSet, goimap.AddFlags, []interface{}{goimap.DeletedFlag}, nil))
		}
		{
			sequenceSet, _ := goimap.ParseSeqSet("3")
			require.NoError(t, client.Store(sequenceSet, goimap.AddFlags, []interface{}{goimap.DeletedFlag}, nil))
		}
		require.NoError(t, client.Close())

		// There are 2 messages in saved-messages.
		mailboxStatus, err := client.Status(messageBoxName, []goimap.StatusItem{goimap.StatusMessages})
		require.NoError(t, err)
		require.Equal(t, uint32(2), mailboxStatus.Messages, "Expected message count does not match")
	})
}
