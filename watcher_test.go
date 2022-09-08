package gluon

import (
	"testing"

	"github.com/ProtonMail/gluon/events"
	"github.com/stretchr/testify/require"
)

func TestWatcher(t *testing.T) {
	watcher := newWatcher(
		events.EventListenerAdded{},
		events.EventListenerRemoved{},
	)

	// The watcher is watching the correct types.
	require.True(t, watcher.isWatching(events.EventListenerAdded{}))
	require.True(t, watcher.isWatching(events.EventListenerRemoved{}))

	// The watcher is not watching the incorrect types.
	require.False(t, watcher.isWatching(events.EventLogin{}))
	require.False(t, watcher.isWatching(events.EventSelect{}))

	// Get a channel to read from the watcher.
	resCh := watcher.getChannel()

	// Send some events to the watcher.
	require.True(t, watcher.send(events.EventListenerAdded{}))
	require.True(t, watcher.send(events.EventListenerRemoved{}))

	// Check we can read the events off the channel.
	require.Equal(t, events.EventListenerAdded{}, <-resCh)
	require.Equal(t, events.EventListenerRemoved{}, <-resCh)

	// Close the watcher.
	watcher.close()

	// Sending more events after the watcher is closed should return false.
	require.False(t, watcher.send(events.EventListenerAdded{}))
	require.False(t, watcher.send(events.EventListenerRemoved{}))
}
