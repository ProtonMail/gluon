// Package pchan implements an async buffered priority channel.
package pchan

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/bradenaw/juniper/xslices"
	"golang.org/x/exp/slices"
)

// PChan is an async, priority channel of items of type T.
type PChan[T any] struct {
	items items[T]
	ready chan struct{}
	done  chan struct{}
	lock  sync.Mutex
	wg    sync.WaitGroup
}

type item[T any] struct {
	val  T
	prio int
	done chan struct{}
}

type items[T any] []*item[T]

// New constructs a new PChan which holds items of type T.
func New[T any]() *PChan[T] {
	return &PChan[T]{
		ready: make(chan struct{}),
		done:  make(chan struct{}),
	}
}

// Push pushes the given value onto the channel, optionally with the given priority.
// If no priority is provided, the item receives a priority lower than all other items in the channel.
// The returned channel is closed when the item has been popped off the channel.
func (ch *PChan[T]) Push(val T, withPrio ...int) chan struct{} {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	select {
	case <-ch.done:
		panic("channel is closed")

	default: // ...
	}

	var prio int

	if len(withPrio) > 0 {
		prio = withPrio[0]
	} else if len(ch.items) > 0 {
		prio = ch.items[len(ch.items)-1].prio - 1
	}

	done := make(chan struct{})

	ch.items = slices.Insert(ch.items, ch.getPosition(prio), &item[T]{
		val:  val,
		prio: prio,
		done: done,
	})

	go func() { ch.ready <- struct{}{} }()

	ch.wg.Add(1)

	return done
}

// Pop blocks until an item is available, then returns that item.
// If the channel is already closed, the call returns immediately and the bool value is false.
func (ch *PChan[T]) Pop() (t T, ok bool) {
	select {
	case <-ch.ready: // ...

	case <-ch.done: // ...
	}

	ch.lock.Lock()
	defer ch.lock.Unlock()

	if len(ch.items) == 0 {
		return t, false
	}

	var item *item[T]

	item, ch.items = ch.items[0], ch.items[1:]

	defer close(item.done)

	ch.wg.Done()

	return item.val, true
}

// Peek returns the highest priority item, if any.
// The bool is true if an item was available.
func (ch *PChan[T]) Peek() (t T, ok bool) {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	if len(ch.items) == 0 {
		return t, false
	}

	return ch.items[0].val, true
}

// Len returns the number of items queued.
func (ch *PChan[T]) Len() int {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	return len(ch.items)
}

// Range repeatedly calls the callback with items as they are pushed onto the channel.
// It stops when the channel is closed.
func (ch *PChan[T]) Range(fn func(T)) {
	for {
		val, ok := ch.Pop()
		if !ok {
			return
		}

		fn(val)
	}
}

// Apply applies the given function to all items in the channel.
// The channel is otherwise unmodified.
func (ch *PChan[T]) Apply(fn func(T)) {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	for _, item := range ch.items {
		fn(item.val)
	}
}

// Wait blocks until the queue is empty.
func (ch *PChan[T]) Wait() {
	ch.wg.Wait()
}

// Close closes the channel, returning whatever was still queued on the channel.
func (ch *PChan[T]) Close() []T {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	select {
	case <-ch.done:
		panic("channel is closed")

	default: // ...
	}

	go func() {
		for range ch.items {
			<-ch.ready
		}
	}()

	close(ch.done)

	return xslices.Map(ch.items, func(item *item[T]) T {
		return item.val
	})
}

// String returns a textual representation of the channel.
func (ch *PChan[T]) String() string {
	var res string

	for _, item := range ch.items {
		res += fmt.Sprintf("[%v %v], ", item.val, item.prio)
	}

	return res
}

//nolint:gosec
func (ch *PChan[T]) getPosition(prio int) int {
	lo := slices.IndexFunc(ch.items, func(item *item[T]) bool {
		return item.prio <= prio
	})

	if lo < 0 {
		lo = len(ch.items)
	}

	hi := slices.IndexFunc(ch.items, func(item *item[T]) bool {
		return item.prio < prio
	})

	if hi < 0 {
		hi = len(ch.items)
	}

	if lo == hi {
		return lo
	}

	return rand.Intn(hi-lo+1) + lo
}
