package queue

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestQueuedChannel(t *testing.T) {
	defer goleak.VerifyNone(t)

	// Create a new queued channel.
	queue := NewQueuedChannel[int](3, 3)

	// Push some items to the queue.
	require.True(t, queue.Enqueue(1, 2, 3))

	// Get a go channel to read from the queue.
	resCh := queue.GetChannel()

	// Check we can initially read items off the channel.
	require.Equal(t, 1, <-resCh)
	require.Equal(t, 2, <-resCh)
	require.Equal(t, 3, <-resCh)

	// Push some more items to the queue.
	require.True(t, queue.Enqueue(4, 5, 6))

	// Close the queue before reading the items.
	queue.Close()

	// Check we can still read the three items.
	require.Equal(t, 4, <-resCh)
	require.Equal(t, 5, <-resCh)
	require.Equal(t, 6, <-resCh)

	// Enqueuing more items after the queue is closed should return false.
	require.False(t, queue.Enqueue(7, 8, 9))
}

func TestQueuedChannelDoesNotLeakIfThereAreNoReadersOnCloseAndDiscard(t *testing.T) {
	defer goleak.VerifyNone(t)

	// Create a new queued channel.
	queue := NewQueuedChannel[int](1, 3)

	// Push some items to the queue.
	require.True(t, queue.Enqueue(1, 2, 3))

	queue.CloseAndDiscardQueued()
}
