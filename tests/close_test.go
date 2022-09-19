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
	runManyToOneTestWithAuth(t, defaultServerOptions(t), []int{1, 2}, func(c map[int]*testConnection, _ *testSession) {
		c[1].C("b001 CREATE saved-messages")
		c[1].S("b001 OK CREATE")

		c[1].doAppend(`saved-messages`, `To: 1@pm.me`, `\Seen`).expect("OK")
		c[1].doAppend(`saved-messages`, `To: 2@pm.me`).expect("OK")
		c[1].doAppend(`saved-messages`, `To: 3@pm.me`, `\Seen`).expect("OK")
		c[1].doAppend(`saved-messages`, `To: 4@pm.me`).expect("OK")
		c[1].doAppend(`saved-messages`, `To: 5@pm.me`, `\Seen`).expect("OK")

		c[1].C(`A002 SELECT saved-messages`)
		c[1].Se(`A002 OK [READ-WRITE] SELECT`)

		c[2].C(`B001 SELECT saved-messages`)
		c[2].Se(`B001 OK [READ-WRITE] SELECT`)
		c[2].C(`B002 FETCH 1:* (UID FLAGS)`)
		c[2].S(
			`* 1 FETCH (UID 1 FLAGS (\Seen))`,
			`* 2 FETCH (UID 2 FLAGS ())`,
			`* 3 FETCH (UID 3 FLAGS (\Seen))`,
			`* 4 FETCH (UID 4 FLAGS ())`,
			`* 5 FETCH (UID 5 FLAGS (\Seen))`,
		)
		c[2].OK("B002")

		// TODO: Match flags in any order.
		c[1].C(`A003 STORE 1 +FLAGS (\Deleted)`)
		c[1].S(`* 1 FETCH (FLAGS (\Deleted \Recent \Seen))`)
		c[1].Sx("A003 OK.*")

		c[1].C(`A004 STORE 2 +FLAGS (\Deleted)`)
		c[1].S(`* 2 FETCH (FLAGS (\Deleted \Recent))`)
		c[1].Sx("A004 OK.*")

		c[1].C(`A005 STORE 3 +FLAGS (\Deleted)`)
		c[1].S(`* 3 FETCH (FLAGS (\Deleted \Recent \Seen))`)
		c[1].Sx("A005 OK.*")

		c[2].C("B003 NOOP")
		c[2].S(
			`* 1 FETCH (FLAGS (\Deleted \Seen))`,
			`* 2 FETCH (FLAGS (\Deleted))`,
			`* 3 FETCH (FLAGS (\Deleted \Seen))`,
		)
		c[2].OK("B003")

		c[2].C("B004 UID EXPUNGE 1")
		c[2].Se(`* 1 EXPUNGE`)
		c[2].OK("B004")

		c[1].C(`A202 CLOSE`)
		c[1].S("A202 OK CLOSE")

		c[2].C("B003 NOOP")
		c[2].S(
			`* 1 EXPUNGE`,
			`* 1 EXPUNGE`,
		)
		c[2].OK("B003")

		// There are 2 messages in saved-messages.
		c[1].C(`A006 STATUS saved-messages (MESSAGES)`)
		c[1].S(`* STATUS "saved-messages" (MESSAGES 2)`)
		c[1].S(`A006 OK STATUS`)
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
