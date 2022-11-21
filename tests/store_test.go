package tests

import (
	"fmt"
	"testing"

	goimap "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C("b001 CREATE saved-messages")
		c.S("b001 OK CREATE")

		c.doAppend(`saved-messages`, `To: 1@pm.me`, `\Seen`).expect("OK")
		c.doAppend(`saved-messages`, `To: 2@pm.me`).expect("OK")
		c.doAppend(`saved-messages`, `To: 3@pm.me`, `\Seen`).expect("OK")

		c.C(`A002 SELECT saved-messages`)
		c.Se(`A002 OK [READ-WRITE] SELECT`)

		// TODO: Match flags in any order.
		c.C(`A005 FETCH 1:* (FLAGS)`)
		c.S(`* 1 FETCH (FLAGS (\Recent \Seen))`,
			`* 2 FETCH (FLAGS (\Recent))`,
			`* 3 FETCH (FLAGS (\Recent \Seen))`)
		c.OK(`A005`)

		// Add \Deleted and \Draft to the first message.
		c.C(`A003 STORE 1 +FLAGS (\Deleted \Draft)`)
		c.S(`* 1 FETCH (FLAGS (\Deleted \Draft \Recent \Seen))`)
		c.OK(`A003`)
		c.C(`A005 FETCH 1 (FLAGS)`)
		c.S(`* 1 FETCH (FLAGS (\Deleted \Draft \Recent \Seen))`)
		c.OK(`A005`)

		// Remove \Seen from the second message (which does not have it)
		c.C(`A003 STORE 2 -FLAGS (\Seen)`)
		c.OK(`A003`)
		c.C(`A005 FETCH 2 (FLAGS)`)
		c.S(`* 2 FETCH (FLAGS (\Recent))`)
		c.Sx(`A005`)

		// Replace the third message's flags with \Flagged and \Answered.
		c.C(`A003 STORE 3 FLAGS (\Flagged \Answered)`)
		c.S(`* 3 FETCH (FLAGS (\Answered \Flagged \Recent))`)
		c.OK(`A003`)
		c.C(`A005 FETCH 3 (FLAGS)`)
		c.S(`* 3 FETCH (FLAGS (\Answered \Flagged \Recent))`)
		c.OK(`A005`)

		// Any attempt to alter the \Recent flags will fail
		c.C(`A003 STORE 3 FLAGS (\Recent Test)`).BAD(`A003`)
		c.C(`A003 STORE 3 +FLAGS (\RECENT Test)`).BAD(`A003`)
		c.C(`A003 STORE 3 -FLAGS (\RECent)`).BAD(`A003`)
	})
}

func TestStoreSilent(t *testing.T) {
	runManyToOneTestWithAuth(t, defaultServerOptions(t), []int{1, 2}, func(c map[int]*testConnection, s *testSession) {
		// one message in INBOX
		c[1].doAppend(`INBOX`, `To: 1@pm.me`).expect("OK")

		// 2 sessions with INBOX selected
		// Message is only recent in the first.
		for i := 1; i <= 2; i++ {
			c[i].C(`A001 SELECT INBOX`)
			c[i].OK(`A001`)
		}

		// FLAGS. Both sessions get the untagged FETCH response
		c[1].C(`A002 STORE 1 FLAGS (flag1 flag2)`)
		c[1].S(`* 1 FETCH (FLAGS (\Recent flag1 flag2))`)
		c[1].OK(`A002`)
		c[2].C(`B002 NOOP`)
		c[2].S(`* 1 FETCH (FLAGS (flag1 flag2))`)
		c[2].OK(`B002`)

		// +FLAGS. Both sessions get the untagged FETCH response
		c[1].C(`A003 STORE 1 +FLAGS (flag3)`)
		c[1].S(`* 1 FETCH (FLAGS (\Recent flag1 flag2 flag3))`)
		c[1].OK(`A003`)
		c[2].C(`B003 NOOP`)
		c[2].S(`* 1 FETCH (FLAGS (flag1 flag2 flag3))`)
		c[2].OK(`B003`)

		// -FLAGS. Both sessions get the untagged FETCH response
		c[1].C(`A004 STORE 1 -FLAGS (flag3 flag2)`)
		c[1].S(`* 1 FETCH (FLAGS (\Recent flag1))`)
		c[1].OK(`A004`)
		c[2].C(`B004 NOOP`)
		c[2].S(`* 1 FETCH (FLAGS (flag1))`)
		c[2].OK(`B004`)

		// FLAGS.SILENT Only session 2 gets the untagged FETCH response
		c[1].C(`A005 STORE 1 FLAGS.SILENT (flag1 flag2)`)
		c[1].OK(`A005`)
		c[2].C(`B005 NOOP`)
		c[2].S(`* 1 FETCH (FLAGS (flag1 flag2))`)
		c[2].OK(`B005`)

		// +FLAGS.SILENT Only session 2 gets the untagged FETCH response
		c[1].C(`A006 STORE 1 +FLAGS.SILENT (flag3)`)
		c[1].OK(`A006`)
		c[2].C(`B006 NOOP`)
		c[2].S(`* 1 FETCH (FLAGS (flag1 flag2 flag3))`)
		c[2].OK(`B006`)

		// -FLAGS.SILENT Only session 2 gets the untagged FETCH response
		c[1].C(`A007 STORE 1 -FLAGS.SILENT (flag3 flag2)`)
		c[1].OK(`A007`)
		c[2].C(`B007 NOOP`)
		c[2].S(`* 1 FETCH (FLAGS (flag1))`)
		c[2].OK(`B007`)
	})
}

func TestStoreReadUnread(t *testing.T) {
	runManyToOneTestWithAuth(t, defaultServerOptions(t), []int{1, 2}, func(c map[int]*testConnection, _ *testSession) {
		c[1].C(`tag select inbox`).OK(`tag`)
		c[2].C(`tag select inbox`).OK(`tag`)

		for i := 1; i <= 10; i++ {
			c[1].doAppend(`INBOX`, fmt.Sprintf(`To: %v@pm.me`, i)).expect("OK")

			c[1].C(`tag noop`).OK(`tag`)
			c[2].C(`tag noop`).OK(`tag`)
		}

		for i := 1; i <= 10; i++ {
			// Begin IDLEing on session 2.
			c[2].C(`tag idle`).Continue()

			// Mark the message as read with session 1.
			c[1].Cf(`tag UID STORE %v +FLAGS.SILENT (\Seen)`, i).OK(`tag`)

			// Mark the message as read with session 1.
			c[1].Cf(`tag UID STORE %v FLAGS.SILENT (\Seen)`, i).OK(`tag`)

			// Wait for the untagged FETCH response on session 2.
			c[2].S(fmt.Sprintf(`* %v FETCH (FLAGS (\Seen))`, i))

			// End IDLEing on session 2.
			c[2].C(`DONE`).OK(`tag`)

			// Both sessions should see the message as read.
			c[1].Cf(`tag UID FETCH %v (FLAGS)`, i).Sxe(`\Seen`).OK(`tag`)
			c[2].Cf(`tag UID FETCH %v (FLAGS)`, i).Sxe(`\Seen`).OK(`tag`)
		}
	})
}

func TestUIDStore(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C("b001 CREATE saved-messages")
		c.S("b001 OK CREATE")

		c.doAppend(`saved-messages`, `To: 1@pm.me`, `\Seen`).expect("OK")
		c.doAppend(`saved-messages`, `To: 2@pm.me`).expect("OK")
		c.doAppend(`saved-messages`, `To: 3@pm.me`, `\Seen`).expect("OK")

		c.C(`A002 SELECT saved-messages`)
		c.Se(`A002 OK [READ-WRITE] SELECT`)

		// TODO: Match flags in any order.
		c.C(`A005 FETCH 1:* (FLAGS)`)
		c.S(`* 1 FETCH (FLAGS (\Recent \Seen))`,
			`* 2 FETCH (FLAGS (\Recent))`,
			`* 3 FETCH (FLAGS (\Recent \Seen))`)
		c.OK(`A005`)

		// Add \Deleted and \Draft to the first message.
		c.C(`A003 UID STORE 1 +FLAGS (\Deleted \Draft)`)
		c.S(`* 1 FETCH (FLAGS (\Deleted \Draft \Recent \Seen) UID 1)`)
		c.OK(`A003`)
		c.C(`A005 UID FETCH 1 (FLAGS)`)
		c.S(`* 1 FETCH (FLAGS (\Deleted \Draft \Recent \Seen) UID 1)`)
		c.OK(`A005`)

		// Remove \Seen from the second message.
		c.C(`A003 UID STORE 2 -FLAGS (\Seen)`)
		c.OK(`A003`)
		c.C(`A005 UID FETCH 2 (FLAGS)`)
		c.S(`* 2 FETCH (FLAGS (\Recent) UID 2)`)
		c.OK(`A005`)

		// Replace the third message's flags with \Flagged and \Answered.
		c.C(`A003 UID STORE 3 FLAGS (\Flagged \Answered)`)
		c.S(`* 3 FETCH (FLAGS (\Answered \Flagged \Recent) UID 3)`)
		c.OK(`A003`)
		c.C(`A005 UID FETCH 3 (FLAGS)`)
		c.S(`* 3 FETCH (FLAGS (\Answered \Flagged \Recent) UID 3)`)
		c.OK(`A005`)

		// Any attempt to alter the \Recent flags will fail
		c.C(`A003 UID STORE 3 FLAGS (\Recent Test)`).BAD(`A003`)
		c.C(`A003 UID STORE 3 +FLAGS (\RECENT Test)`).BAD(`A003`)
		c.C(`A003 UID STORE 3 -FLAGS (\RECent)`).BAD(`A003`)
	})
}

func TestFlagsDuplicateAndCaseInsensitive(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.doAppend(`INBOX`, `To: 1@pm.me`).expect("OK")

		c.C(`A001 SELECT INBOX`)
		c.Se(`A001 OK [READ-WRITE] SELECT`)

		// no duplicates
		c.C(`A002 STORE 1 FLAGS (flag1 flag1)`)
		c.S(`* 1 FETCH (FLAGS (\Recent flag1))`)
		c.OK(`A002`)

		// no duplicates and case-insensitive
		c.C(`A003 STORE 1 FLAGS (FLAG1 flag2 FLAG2)`)
		c.S(`* 1 FETCH (FLAGS (FLAG1 \Recent flag2))`)
		c.OK(`A003`)

		// unchanged, no untagged response
		c.C(`A004 STORE 1 FLAGS (FLAG1 flag2 FLAG2)`)
		c.OK(`A004`)
		c.C(`A005 FETCH 1 (FLAGS)`)
		c.S(`* 1 FETCH (FLAGS (FLAG1 \Recent flag2))`)
		c.OK(`A005`)

		// +FLAGS with no changes, no untagged response
		c.C(`A006 STORE 1 +FLAGS (flag1 flag2)`)
		c.OK(`A006`)

		// +FLAGS with a new flag
		c.C(`A007 STORE 1 +FLAGS (FLAG1 FLAG3)`)
		c.S(`* 1 FETCH (FLAGS (FLAG1 FLAG3 \Recent flag2))`)
		c.OK(`A007`)

		// +FLAGS with a empty flag list, no untagged response
		c.C(`A008 STORE 1 +FLAGS ()`)
		c.OK(`A008`)

		// -FLAGS with difference case
		c.C(`A009 STORE 1 -FLAGS (flag3)`)
		c.S(`* 1 FETCH (FLAGS (FLAG1 \Recent flag2))`)
		c.OK(`A009`)

		// -FLAGS -with non-existing flag, no untagged
		c.C(`A00A STORE 1 -FLAGS (flag3)`)
		c.OK(`A00A`)

		// -FLAGS with empty list
		c.C(`A00B STORE 1 -FLAGS ()`)
		c.OK(`A00B`)
	})
}

func TestStoreFlagsPersistBetweenRuns(t *testing.T) {
	options := defaultServerOptions(t, withDataDir(t.TempDir()))

	runOneToOneTestWithAuth(t, options, func(c *testConnection, _ *testSession) {
		c.C("b001 CREATE saved-messages")
		c.S("b001 OK CREATE")
		c.doAppend(`saved-messages`, `To: 2@pm.me`).expect("OK")
	})

	// Check if recent flag was persisted and then mark the message as deleted.
	runOneToOneTestClientWithAuth(t, options, func(client *client.Client, _ *testSession) {
		_, err := client.Select("saved-messages", false)
		require.NoError(t, err)
		newFetchCommand(t, client).withItems(goimap.FetchFlags).fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.wantFlags(goimap.RecentFlag)
		}).check()

		require.NoError(t, client.Store(createSeqSet("1"), goimap.AddFlags, []interface{}{goimap.DeletedFlag}, nil))
	})

	// Check if delete flag was persisted.
	runOneToOneTestClientWithAuth(t, options, func(client *client.Client, _ *testSession) {
		_, err := client.Select("saved-messages", false)
		require.NoError(t, err)
		newFetchCommand(t, client).withItems(goimap.FetchFlags).fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.wantFlags(goimap.DeletedFlag)
		}).check()
	})
}
