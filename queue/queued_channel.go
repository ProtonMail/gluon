package queue

import (
	"sync"
)

// QueuedChannel represents a channel on which queued items can be published without having to worry if the reader
// has actually consumed existing items first or if there's no way of knowing ahead of time what the ideal channel
// buffer size should be.
type QueuedChannel[T any] struct {
	ch     chan T
	items  []T
	cond   *sync.Cond
	closed atomicBool // Should use atomic.Bool once we use Go 1.19!
}

func NewQueuedChannel[T any](chanBufferSize, queueCapacity int) *QueuedChannel[T] {
	queue := &QueuedChannel[T]{
		ch:    make(chan T, chanBufferSize),
		items: make([]T, 0, queueCapacity),
		cond:  sync.NewCond(&sync.Mutex{}),
	}

	// The queue is initially not closed.
	queue.closed.store(false)

	go func() {
		defer close(queue.ch)

		for {
			item, ok := queue.pop()
			if !ok {
				return
			}

			queue.ch <- item
		}
	}()

	return queue
}

func (q *QueuedChannel[T]) Enqueue(items ...T) bool {
	if q.closed.load() {
		return false
	}

	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	q.items = append(q.items, items...)

	q.cond.Broadcast()

	return true
}

func (q *QueuedChannel[T]) GetChannel() <-chan T {
	return q.ch
}

func (q *QueuedChannel[T]) Close() {
	q.closed.store(true)

	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	q.cond.Broadcast()
}

func (q *QueuedChannel[T]) pop() (T, bool) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	var item T

	// Wait until there are items to pop, returning false immediately if the queue is closed.
	// This allows the queue to continue popping elements if it's closed,
	// but will prevent it from hanging indefinitely once it runs out of items.
	for len(q.items) == 0 {
		if q.closed.load() {
			return item, false
		}

		q.cond.Wait()
	}

	item, q.items = q.items[0], q.items[1:]

	return item, true
}
