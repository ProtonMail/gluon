package tests

import (
	"testing"

	"github.com/ProtonMail/gluon/imap"
	"github.com/emersion/go-imap/client"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
)

func TestCapability(t *testing.T) {
	runOneToOneTestClient(t, "user", "pass", "/", func(client *client.Client, s *testSession) {
		capabilities, err := client.Capability()
		require.NoError(t, err)
		require.ElementsMatch(t, maps.Keys(capabilities), []string{
			string(imap.IMAP4rev1),
			string(imap.StartTLS),
			string(imap.IDLE),
			string(imap.UNSELECT),
			string(imap.UIDPLUS),
			string(imap.MOVE),
		})
	})
}
