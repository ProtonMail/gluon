package tests

import (
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/limits"
	goimap "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func testIMAPLimits() limits.IMAP {
	return limits.NewIMAPLimits(
		3, // Needs to be at least 3 due to INBOX and Recovered Messages.
		1,
		2, // Set to 2 so we can create a least on mailbox, delete it and then create a second to trigger the UID check.
		4, // Set to 4 o we can add at least one message, delete it and then add a second message to trigger the UID check.
	)
}

func TestMaxMailboxLimitRespected(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t, withIMAPLimits(testIMAPLimits()), withUIDValidityGenerator(imap.NewIncrementalUIDValidityGenerator())), func(client *client.Client, session *testSession) {
		require.NoError(t, client.Create("mbox1"))
		require.Error(t, client.Create("mbox2"))
	})
}

func TestMaxMessageLimitRespected_Append(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t, withIMAPLimits(testIMAPLimits()), withUIDValidityGenerator(imap.NewIncrementalUIDValidityGenerator())), func(client *client.Client, session *testSession) {
		require.NoError(t, doAppendWithClient(client, "INBOX", "To: Foo@bar.com", time.Now()))
		require.Error(t, doAppendWithClient(client, "INBOX", "To: Bar@bar.com", time.Now()))
	})
}

func TestMaxUIDLimitRespected_Append(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t, withIMAPLimits(testIMAPLimits()), withUIDValidityGenerator(imap.NewIncrementalUIDValidityGenerator())), func(client *client.Client, session *testSession) {
		require.NoError(t, doAppendWithClient(client, "INBOX", "To: Foo@bar.com", time.Now()))
		_, err := client.Select("INBOX", false)
		require.NoError(t, err)
		require.NoError(t, client.Store(createSeqSet("1"), goimap.AddFlags, []interface{}{goimap.DeletedFlag}, nil))
		require.NoError(t, client.Expunge(nil))
		// Append should fail now as we triggered max UID validity error.
		require.Error(t, doAppendWithClient(client, "INBOX", "To: Bar@bar.com", time.Now()))
	})
}

func TestMaxMessageLimitRespected_Copy(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t, withIMAPLimits(testIMAPLimits()), withUIDValidityGenerator(imap.NewIncrementalUIDValidityGenerator())), func(client *client.Client, session *testSession) {
		session.setUpdatesAllowedToFail("user", true)
		require.NoError(t, client.Create("mbox1"))
		require.NoError(t, doAppendWithClient(client, "mbox1", "To: Foo@bar.com", time.Now()))
		require.NoError(t, doAppendWithClient(client, "INBOX", "To: Bar@bar.com", time.Now()))
		_, err := client.Select("INBOX", false)
		require.NoError(t, err)
		require.Error(t, client.Copy(createSeqSet("1"), "mbox1"))
	})
}

func TestMaxUIDLimitRespected_Copy(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t, withIMAPLimits(testIMAPLimits()), withUIDValidityGenerator(imap.NewIncrementalUIDValidityGenerator())), func(client *client.Client, session *testSession) {
		session.setUpdatesAllowedToFail("user", true)
		require.NoError(t, client.Create("mbox1"))
		require.NoError(t, doAppendWithClient(client, "mbox1", "To: Foo@bar.com", time.Now()))
		require.NoError(t, doAppendWithClient(client, "INBOX", "To: Bar@bar.com", time.Now()))

		// Delete existing message in mbox1 to trigget UID validity check
		_, err := client.Select("mbox1", false)
		require.NoError(t, err)
		require.NoError(t, client.Store(createSeqSet("1"), goimap.AddFlags, []interface{}{goimap.DeletedFlag}, nil))
		require.NoError(t, client.Expunge(nil))

		// Try to copy message to mbox
		_, err = client.Select("INBOX", false)
		require.NoError(t, err)
		require.Error(t, client.Copy(createSeqSet("1"), "mbox1"))
	})
}

func TestMaxMessageLimitRespected_Move(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t, withIMAPLimits(testIMAPLimits()), withUIDValidityGenerator(imap.NewIncrementalUIDValidityGenerator())), func(client *client.Client, session *testSession) {
		session.setUpdatesAllowedToFail("user", true)
		require.NoError(t, client.Create("mbox1"))
		require.NoError(t, doAppendWithClient(client, "mbox1", "To: Foo@bar.com", time.Now()))
		require.NoError(t, doAppendWithClient(client, "INBOX", "To: Bar@bar.com", time.Now()))
		_, err := client.Select("INBOX", false)
		require.NoError(t, err)
		require.Error(t, client.Move(createSeqSet("1"), "mbox1"))
	})
}

func TestMaxUIDLimitRespected_Move(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t, withIMAPLimits(testIMAPLimits()), withUIDValidityGenerator(imap.NewIncrementalUIDValidityGenerator())), func(client *client.Client, session *testSession) {
		session.setUpdatesAllowedToFail("user", true)
		require.NoError(t, client.Create("mbox1"))
		require.NoError(t, doAppendWithClient(client, "mbox1", "To: Foo@bar.com", time.Now()))
		require.NoError(t, doAppendWithClient(client, "INBOX", "To: Bar@bar.com", time.Now()))

		// Delete existing message in mbox1 to trigget UID validity check
		_, err := client.Select("mbox1", false)
		require.NoError(t, err)
		require.NoError(t, client.Store(createSeqSet("1"), goimap.AddFlags, []interface{}{goimap.DeletedFlag}, nil))
		require.NoError(t, client.Expunge(nil))

		// Try to copy message to mbox
		_, err = client.Select("INBOX", false)
		require.NoError(t, err)
		require.Error(t, client.Move(createSeqSet("1"), "mbox1"))
	})
}

func TestMaxUIDValidityLimitRespected(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t, withIMAPLimits(testIMAPLimits()), withUIDValidityGenerator(imap.NewIncrementalUIDValidityGenerator())), func(client *client.Client, session *testSession) {
		session.setUpdatesAllowedToFail("user", true)
		require.NoError(t, client.Create("mbox1"))
		require.NoError(t, client.Delete("mbox1"))
		require.Error(t, client.Create("mbox2"))
	})
}
