package store

import (
	"sync"
	"testing"

	"github.com/ProtonMail/gluon/async"
	"github.com/stretchr/testify/assert"
)

func TestSemaphore(t *testing.T) {
	sem := NewSemaphore(4, async.NoopPanicHandler{})

	// Block the semaphore so that tasks wait until we unblock it.
	sem.Block()

	// Collect results in a list.
	var res list[int]

	// Start some jobs; they won't execute yet because the semaphore is blocked.
	wait := wait(func() {
		sem.Go(func() { res.insert(1) })
		sem.Go(func() { res.insert(2) })
		sem.Go(func() { res.insert(3) })
		sem.Go(func() { res.insert(4) })
	})

	// The semaphore is blocked so none of the tasks should have executed yet.
	assert.Equal(t, []int{}, res.items())

	// Unblock the semaphore so that blocked tasks can execute.
	sem.Unblock()

	// Wait for the jobs to be given to the semaphore.
	wait()

	// Tell the semaphore to wait for its tasks to finish.
	sem.Wait()

	// The jobs should have executed now.
	assert.ElementsMatch(t, []int{1, 2, 3, 4}, res.items())
}

func wait(fn func()) func() {
	var wg sync.WaitGroup

	wg.Add(1)

	go func() { defer wg.Done(); fn() }()

	return wg.Wait
}

type list[T any] struct {
	val []T
	mut sync.Mutex
}

func (l *list[T]) insert(val T) {
	l.mut.Lock()
	defer l.mut.Unlock()

	l.val = append(l.val, val)
}

func (l *list[T]) items() []T {
	l.mut.Lock()
	defer l.mut.Unlock()

	items := append([]T{}, l.val...)

	return items
}
