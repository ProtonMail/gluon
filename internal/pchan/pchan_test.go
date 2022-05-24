package pchan

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
)

func TestPChanPush(t *testing.T) {
	ch := New[int]()
	defer ch.Close()

	wantVals := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	// Push all with the same priority so that their order is randomized.
	for _, val := range wantVals {
		ch.Push(val, 1)
	}

	var haveVals []int

	for range wantVals {
		haveVal, ok := ch.Pop()
		if !ok {
			break
		}

		haveVals = append(haveVals, haveVal)
	}

	assert.ElementsMatch(t, haveVals, wantVals)
}

func TestPChanPushOrdered(t *testing.T) {
	ch := New[int]()
	defer ch.Close()

	wantVals := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	// Push all with no priority so their order is constant.
	for _, val := range wantVals {
		ch.Push(val)
	}

	var haveVals []int

	for range wantVals {
		haveVal, ok := ch.Pop()
		if !ok {
			break
		}

		haveVals = append(haveVals, haveVal)
	}

	assert.Equal(t, haveVals, wantVals)
}

func TestPChanPushClosed(t *testing.T) {
	ch := New[int]()

	assert.NotPanics(t, func() { ch.Push(1) })
	assert.NotPanics(t, func() { ch.Push(2) })
	assert.NotPanics(t, func() { ch.Push(3) })

	ch.Close()

	assert.Panics(t, func() { ch.Push(4) })
	assert.Panics(t, func() { ch.Push(5) })
	assert.Panics(t, func() { ch.Push(6) })
}

func TestPChanPushConcurrent(t *testing.T) {
	ch := New[string]()
	defer ch.Close()

	var wg sync.WaitGroup

	// We are going to test with 6 additional goroutines.
	wg.Add(6)

	// Start many concurrent pushes.
	go func() { defer wg.Done(); ch.Push("a", 1) }()
	go func() { defer wg.Done(); ch.Push("a", 1) }()
	go func() { defer wg.Done(); ch.Push("b", 2) }()
	go func() { defer wg.Done(); ch.Push("b", 2) }()
	go func() { defer wg.Done(); ch.Push("c", 3) }()
	go func() { defer wg.Done(); ch.Push("c", 3) }()

	// Wait for the items to be pushed.
	wg.Wait()

	// All 6 should now be ready for popping.
	require.Len(t, ch.items, 6)

	// They should be popped in priority order.
	assert.Equal(t, "c", getValue(t, ch))
	assert.Equal(t, "c", getValue(t, ch))
	assert.Equal(t, "b", getValue(t, ch))
	assert.Equal(t, "b", getValue(t, ch))
	assert.Equal(t, "a", getValue(t, ch))
	assert.Equal(t, "a", getValue(t, ch))
}

func TestPChanPopClosed(t *testing.T) {
	ch := New[int]()

	ch.Push(1)
	ch.Push(2)
	ch.Push(3)
	ch.Close()

	val1, ok1 := ch.Pop()
	val2, ok2 := ch.Pop()
	val3, ok3 := ch.Pop()
	valClose, okClose := ch.Pop()

	assert.Equal(t, []int{1, 2, 3, *new(int)}, []int{val1, val2, val3, valClose})
	assert.Equal(t, []bool{true, true, true, false}, []bool{ok1, ok2, ok3, okClose})
}

func TestPChanPopConcurrent(t *testing.T) {
	ch := New[int]()
	defer ch.Close()

	var (
		wg   sync.WaitGroup
		res  []int
		lock sync.Mutex
	)

	// We are going to test with 5 additional goroutines.
	wg.Add(5)

	// Start 5 concurrent pops; these consume any items pushed.
	go func() { defer wg.Done(); lock.Lock(); defer lock.Unlock(); res = append(res, getValue(t, ch)) }()
	go func() { defer wg.Done(); lock.Lock(); defer lock.Unlock(); res = append(res, getValue(t, ch)) }()
	go func() { defer wg.Done(); lock.Lock(); defer lock.Unlock(); res = append(res, getValue(t, ch)) }()
	go func() { defer wg.Done(); lock.Lock(); defer lock.Unlock(); res = append(res, getValue(t, ch)) }()
	go func() { defer wg.Done(); lock.Lock(); defer lock.Unlock(); res = append(res, getValue(t, ch)) }()

	// Push and block; items should be popped immediately by the waiting goroutines.
	<-ch.Push(1, 1)
	<-ch.Push(2, 2)
	<-ch.Push(3, 3)
	<-ch.Push(4, 4)
	<-ch.Push(5, 5)

	// Wait for all items to be popped then close the result channel.
	wg.Wait()

	assert.True(t, slices.IsSorted(res))
}

func TestPChanPeek(t *testing.T) {
	ch := New[int]()

	ch.Push(1)
	ch.Push(2)
	ch.Push(3)

	peekVal1, ok := ch.Peek()
	assert.True(t, ok)
	assert.Equal(t, 1, peekVal1)

	popVal1, ok := ch.Pop()
	assert.True(t, ok)
	assert.Equal(t, 1, popVal1)

	peekVal2, ok := ch.Peek()
	assert.True(t, ok)
	assert.Equal(t, 2, peekVal2)

	popVal2, ok := ch.Pop()
	assert.True(t, ok)
	assert.Equal(t, 2, popVal2)

	peekVal3, ok := ch.Peek()
	assert.True(t, ok)
	assert.Equal(t, 3, peekVal3)

	popVal3, ok := ch.Pop()
	assert.True(t, ok)
	assert.Equal(t, 3, popVal3)

	peekVal4, ok := ch.Peek()
	assert.False(t, ok)
	assert.Equal(t, *new(int), peekVal4)
}

func TestPChanRange(t *testing.T) {
	ch := New[string]()

	// Push some items then close the channel.
	go func() {
		<-ch.Push("a", 1)
		<-ch.Push("b", 1)
		<-ch.Push("c", 1)

		ch.Close()
	}()

	var res []string

	// Retrieve the items. The range will stop when the channel is closed.
	ch.Range(func(val string) {
		res = append(res, val)
	})

	assert.ElementsMatch(t, []string{"a", "b", "c"}, res)
}

func TestPChanRangeClosed(t *testing.T) {
	ch := New[int]()

	ch.Push(1)
	ch.Push(2)
	ch.Push(3)
	ch.Close()

	var vals []int

	ch.Range(func(val int) {
		vals = append(vals, val)
	})

	assert.Equal(t, []int{1, 2, 3}, vals)
}

func TestPChanClose(t *testing.T) {
	ch := New[int]()

	ch.Push(1)
	ch.Push(2)
	ch.Push(3)

	// First close does not panic.
	require.Equal(t, []int{1, 2, 3}, ch.Close())

	// Second close does panic.
	assert.Panics(t, func() { ch.Close() })
}

func getValue[T any](t *testing.T, ch *PChan[T]) T {
	val, ok := ch.Pop()

	assert.True(t, ok)

	return val
}
