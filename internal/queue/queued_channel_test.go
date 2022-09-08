package queue

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestQueuedChannel(t *testing.T) {
	queue := NewQueuedChannel[int](3, 3)

	// Push some items to the queue.
	queue.Queue(1, 2, 3)

	// Get a go channel to read from the queue.
	resCh := queue.GetChannel()

	// Close the queue before reading the items.
	queue.Close()

	// Check we can still read the three items.
	require.Equal(t, 1, <-resCh)
	require.Equal(t, 2, <-resCh)
	require.Equal(t, 3, <-resCh)
}
