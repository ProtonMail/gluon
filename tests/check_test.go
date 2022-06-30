package tests

import (
	"testing"

	"github.com/emersion/go-imap/client"
	"github.com/stretchr/testify/require"
)

func TestCheck(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		mailboxStatus, err := client.Select("INBOX", false)
		require.NoError(t, err)
		require.Equal(t, false, mailboxStatus.ReadOnly)
		require.NoError(t, client.Check())
	})
}
