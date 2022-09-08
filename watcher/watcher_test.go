package watcher

import (
	"testing"

	"github.com/ProtonMail/gluon/events"
	"github.com/stretchr/testify/require"
)

func TestWatcher(t *testing.T) {
	watcher := New[events.Event](
		events.EventListenerAdded{},
		events.EventListenerRemoved{},
	)

	// The watcher is watching the correct types.
	require.True(t, watcher.IsWatching(events.EventListenerAdded{}))
	require.True(t, watcher.IsWatching(events.EventListenerRemoved{}))

	// The watcher is not watching the incorrect types.
	require.False(t, watcher.IsWatching(events.EventLogin{}))
	require.False(t, watcher.IsWatching(events.EventSelect{}))

	// Get a channel to read from the watcher.
	resCh := watcher.GetChannel()

	// Send some events to the watcher.
	require.True(t, watcher.Send(events.EventListenerAdded{}))
	require.True(t, watcher.Send(events.EventListenerRemoved{}))

	// Check we can read the events off the channel.
	require.Equal(t, events.EventListenerAdded{}, <-resCh)
	require.Equal(t, events.EventListenerRemoved{}, <-resCh)

	// Close the watcher.
	watcher.Close()

	// Sending more events after the watcher is closed should return false.
	require.False(t, watcher.Send(events.EventListenerAdded{}))
	require.False(t, watcher.Send(events.EventListenerRemoved{}))
}
