package queue

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCTQueue_PushPop(t *testing.T) {
	queue := NewCTQueue[int]()

	const pushValue = 20

	go func() {
		require.True(t, queue.Push(pushValue))
	}()

	v, ok := queue.Pop()
	require.True(t, ok)
	require.Equal(t, v, pushValue)
}

func TestCTQueue_PushCloseAndRetrieve(t *testing.T) {
	queue := NewCTQueue[int]()

	expectedValues := []int{20, 30, 40, 50}

	for _, v := range expectedValues {
		require.False(t, queue.IsClosed())
		require.True(t, queue.Push(v))
	}

	require.Equal(t, len(expectedValues), queue.Len())

	values := queue.CloseAndRetrieveRemaining()
	require.True(t, queue.IsClosed())
	require.ElementsMatch(t, expectedValues, values)

	require.False(t, queue.Push(100))
	require.Zero(t, queue.Len())
}

func TestCTQueue_CloseAndPop(t *testing.T) {
	queue := NewCTQueue[int]()
	expectedValues := []int{20, 30, 40, 50}

	go func() {
		// Sleep for 0.5 sec
		time.Sleep(500 * time.Millisecond)
		require.False(t, queue.IsClosed())
		require.True(t, queue.PushMany(expectedValues...))
		require.Equal(t, len(expectedValues), queue.Len())
		queue.Close()
	}()

	var values []int

	for {
		v, ok := queue.Pop()
		if ok {
			values = append(values, v)
		} else {
			break
		}
	}

	require.ElementsMatch(t, expectedValues, values)
	require.True(t, queue.IsClosed())
	require.Zero(t, queue.Len())
}

func TestCTQueue_CloseWakesBlockedPop(t *testing.T) {
	queue := NewCTQueue[int]()

	go func() {
		// Sleep for 0.5 sec
		time.Sleep(500 * time.Millisecond)
		queue.Close()
	}()

	_, ok := queue.Pop()
	require.False(t, ok)
	require.True(t, queue.IsClosed())
}
