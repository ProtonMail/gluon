package tests

import (
	"strings"
	"testing"
	"time"

	"github.com/ProtonMail/gluon/imap"
	goimap "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/stretchr/testify/require"
)

func TestExamineWithLiteral(t *testing.T) {
	// This test remains here since it is not possible to send literal commands with the
	// IMAP client. The rest of the functionality is still tested in the IMAP client test.
	runOneToOneTestWithAuth(t, defaultServerOptions(t, withUIDValidityGenerator(imap.NewFixedUIDValidityGenerator(imap.UID(1)))), func(c *testConnection, _ *testSession) {
		c.C("A002 CREATE Archive")
		c.S("A002 OK CREATE")

		c.doAppend(`Archive`, `To: 3@pm.me`, `\Seen`).expect("OK")

		c.C("a007 examine {7}")
		c.S("+ Ready")
		c.C("Archive")
		c.S(`* FLAGS (\Deleted \Flagged \Seen)`,
			`* 1 EXISTS`,
			`* 1 RECENT`,
			`* OK [PERMANENTFLAGS (\Deleted \Flagged \Seen)]`,
			`* OK [UIDNEXT 2]`,
			`* OK [UIDVALIDITY 1]`)
		c.S(`a007 OK [READ-ONLY] EXAMINE`)
	})
}

func TestExamineClient(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t, withUIDValidityGenerator(imap.NewFixedUIDValidityGenerator(1))), func(client *client.Client, _ *testSession) {
		require.NoError(t, client.Create("Archive"))
		require.NoError(t, client.Append("INBOX", []string{goimap.SeenFlag}, time.Now(), strings.NewReader("To: 1@pm.me")))
		require.NoError(t, client.Append("INBOX", []string{}, time.Now(), strings.NewReader("To: 2@pm.me")))
		require.NoError(t, client.Append("Archive", []string{goimap.SeenFlag}, time.Now(), strings.NewReader("To: 3@pm.me")))

		// IMAP client does not have an explicit Examine call, but this a call to Select(..., readonly=true) gets
		// converted into an EXAMINE command.
		{
			mailboxStatus, err := client.Select("INBOX", true)
			require.NoError(t, err)
			require.Equal(t, uint32(2), mailboxStatus.Messages)
			require.Equal(t, uint32(2), mailboxStatus.Recent)
			require.Equal(t, uint32(2), mailboxStatus.UnseenSeqNum)
			require.ElementsMatch(t, mailboxStatus.Flags, []string{goimap.SeenFlag, goimap.DeletedFlag, goimap.FlaggedFlag})
			require.ElementsMatch(t, mailboxStatus.PermanentFlags, []string{goimap.SeenFlag, goimap.DeletedFlag, goimap.FlaggedFlag})
			require.Equal(t, uint32(3), mailboxStatus.UidNext)
			require.Equal(t, uint32(1), mailboxStatus.UidValidity)
			require.Equal(t, true, mailboxStatus.ReadOnly)
		}
		// Examining INBOX again DOES NOT modify the RECENT value.
		{
			mailboxStatus, err := client.Select("INBOX", true)
			require.NoError(t, err)
			require.Equal(t, uint32(2), mailboxStatus.Messages)
			require.Equal(t, uint32(2), mailboxStatus.Recent)
			require.Equal(t, uint32(2), mailboxStatus.UnseenSeqNum)
			require.ElementsMatch(t, mailboxStatus.Flags, []string{goimap.SeenFlag, goimap.DeletedFlag, goimap.FlaggedFlag})
			require.ElementsMatch(t, mailboxStatus.PermanentFlags, []string{goimap.SeenFlag, goimap.DeletedFlag, goimap.FlaggedFlag})
			require.Equal(t, uint32(3), mailboxStatus.UidNext)
			require.Equal(t, uint32(1), mailboxStatus.UidValidity)
			require.Equal(t, true, mailboxStatus.ReadOnly)
		}
		{
			mailboxStatus, err := client.Select("Archive", true)
			require.NoError(t, err)
			require.Equal(t, uint32(1), mailboxStatus.Messages)
			require.Equal(t, uint32(1), mailboxStatus.Recent)
			require.Equal(t, uint32(0), mailboxStatus.UnseenSeqNum)
			require.ElementsMatch(t, mailboxStatus.Flags, []string{goimap.SeenFlag, goimap.DeletedFlag, goimap.FlaggedFlag})
			require.ElementsMatch(t, mailboxStatus.PermanentFlags, []string{goimap.SeenFlag, goimap.DeletedFlag, goimap.FlaggedFlag})
			require.Equal(t, uint32(2), mailboxStatus.UidNext)
			require.Equal(t, uint32(1), mailboxStatus.UidValidity)
			require.Equal(t, true, mailboxStatus.ReadOnly)
		}
	})
}
