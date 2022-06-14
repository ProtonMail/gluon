package queue

import (
	"sync"
	"sync/atomic"
)

// CTQueue is a thread safe queue implementation which allows safe usage among different threads/go-routines and can
// be closed at any given time. The queue can be closed by either allowing the remaining elements to remain within or
// to be extracted. Once the queue is closed, no new elements may be added, but existing elements can still be popped of
// the queue.
type CTQueue[T any] struct {
	items  []T
	cond   *sync.Cond
	closed int32
}

func NewCTQueue[T any]() *CTQueue[T] {
	return &CTQueue[T]{
		items:  nil,
		cond:   sync.NewCond(&sync.Mutex{}),
		closed: 0,
	}
}

func (ctq *CTQueue[T]) Push(val T) bool {
	ctq.cond.L.Lock()
	defer ctq.cond.L.Unlock()

	if ctq.IsClosed() {
		return false
	}

	ctq.items = append(ctq.items, val)
	ctq.cond.Broadcast()

	return true
}

func (ctq *CTQueue[T]) PushMany(val ...T) bool {
	ctq.cond.L.Lock()
	defer ctq.cond.L.Unlock()

	if ctq.IsClosed() {
		return false
	}

	ctq.items = append(ctq.items, val...)
	ctq.cond.Broadcast()

	return true
}

func (ctq *CTQueue[T]) Pop() (T, bool) {
	ctq.cond.L.Lock()
	defer ctq.cond.L.Unlock()

	for len(ctq.items) == 0 {
		// Check if the queue has been closed before going to sleep.
		// This allows the queue to continue popping elements if it's closed, but will prevent it from hanging
		// indefinitely once it runs out of items.
		if ctq.IsClosed() {
			var r T
			return r, false
		}

		ctq.cond.Wait()
	}

	var item T
	item, ctq.items = ctq.items[0], ctq.items[1:]

	return item, true
}

// TryPop attempts to pop an element of the queue, but if no elements are available it doesn't wait and exits
//immediately.
func (ctq *CTQueue[T]) TryPop() (T, bool) {
	ctq.cond.L.Lock()
	defer ctq.cond.L.Unlock()

	for len(ctq.items) == 0 || ctq.IsClosed() {
		var r T
		return r, false
	}

	var item T
	item, ctq.items = ctq.items[0], ctq.items[1:]

	return item, true
}

func (ctq *CTQueue[T]) IsClosed() bool {
	return atomic.LoadInt32(&ctq.closed) == 1
}

func (ctq *CTQueue[T]) markClosed() {
	atomic.StoreInt32(&ctq.closed, 1)
}

func (ctq *CTQueue[T]) CloseAndRetrieveRemaining() []T {
	ctq.markClosed()
	ctq.cond.L.Lock()
	defer ctq.cond.L.Unlock()
	items := ctq.items
	ctq.items = nil
	ctq.cond.Broadcast()

	return items
}

func (ctq *CTQueue[T]) Close() {
	ctq.markClosed()
	ctq.cond.L.Lock()
	defer ctq.cond.L.Unlock()
	ctq.cond.Broadcast()
}

func (ctq *CTQueue[T]) Len() int {
	ctq.cond.L.Lock()
	defer ctq.cond.L.Unlock()

	return len(ctq.items)
}

// Apply applies the given function to all items in the queue.
func (ctq *CTQueue[T]) Apply(fn func(T)) {
	ctq.cond.L.Lock()
	defer ctq.cond.L.Unlock()

	for _, item := range ctq.items {
		fn(item)
	}
}
