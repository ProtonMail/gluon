package watcher

import (
	"testing"

	"github.com/ProtonMail/gluon/events"
	"github.com/ProtonMail/gluon/queue"
	"github.com/stretchr/testify/require"
)

func TestWatcher(t *testing.T) {
	watcher := New[events.Event](
		queue.NoopPanicHandler{},
		events.ListenerAdded{},
		events.ListenerRemoved{},
	)

	// The watcher is watching the correct types.
	require.True(t, watcher.IsWatching(events.ListenerAdded{}))
	require.True(t, watcher.IsWatching(events.ListenerRemoved{}))

	// The watcher is not watching the incorrect types.
	require.False(t, watcher.IsWatching(events.Login{}))
	require.False(t, watcher.IsWatching(events.Select{}))

	// Get a channel to read from the watcher.
	resCh := watcher.GetChannel()

	// Send some events to the watcher.
	require.True(t, watcher.Send(events.ListenerAdded{}))
	require.True(t, watcher.Send(events.ListenerRemoved{}))

	// Check we can read the events off the channel.
	require.Equal(t, events.ListenerAdded{}, <-resCh)
	require.Equal(t, events.ListenerRemoved{}, <-resCh)

	// Close the watcher.
	watcher.Close()

	// Sending more events after the watcher is closed should return false.
	require.False(t, watcher.Send(events.ListenerAdded{}))
	require.False(t, watcher.Send(events.ListenerRemoved{}))
}
