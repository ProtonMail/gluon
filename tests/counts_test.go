package tests

import (
	"testing"
	"time"

	"github.com/ProtonMail/gluon/events"
	"github.com/ProtonMail/gluon/internal/ids"
	goimap "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/stretchr/testify/require"
)

func TestCounts(t *testing.T) {
	dir := t.TempDir()

	runOneToOneTestClientWithAuth(t, defaultServerOptions(t, withDataDir(dir)), func(client *client.Client, s *testSession) {
		for _, count := range getEvent[events.UserAdded](s.eventCh).Counts {
			require.Equal(t, 0, count)
		}

		for i := 0; i < 10; i++ {
			require.NoError(t, doAppendWithClientFromFile(t, client, "INBOX", "testdata/afternoon-meeting.eml", time.Now(), goimap.SeenFlag))
		}
	})

	runOneToOneTestClientWithAuth(t, defaultServerOptions(t, withDataDir(dir)), func(_ *client.Client, s *testSession) {
		for mbox, count := range getEvent[events.UserAdded](s.eventCh).Counts {
			if mbox == ids.GluonInternalRecoveryMailboxRemoteID {
				require.Equal(t, 0, count)
			} else {
				require.Equal(t, 10, count)
			}
		}
	})
}

func getEvent[T any](eventCh <-chan events.Event) T {
	for event := range eventCh {
		if event, ok := event.(T); ok {
			return event
		}
	}

	panic("no event")
}
