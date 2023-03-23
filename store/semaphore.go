package store

import "sync"

type PanicHandler interface {
	HandlePanic()
}

// Semaphore implements a type used to limit concurrent operations.
type Semaphore struct {
	ch chan struct{}
	wg sync.WaitGroup
	rw sync.RWMutex

	panicHandler PanicHandler
}

// NewSemaphore constructs a new semaphore with the given limit.
func NewSemaphore(max int) *Semaphore {
	return &Semaphore{ch: make(chan struct{}, max)}
}

// Lock locks the semaphore, waiting first until it is possible.
func (sem *Semaphore) Lock() {
	sem.rw.RLock()
	sem.ch <- struct{}{}
}

// Unlock unlocks the semaphore.
func (sem *Semaphore) Unlock() {
	sem.rw.RUnlock()
	<-sem.ch
}

// Block prevents the semaphore from being locked.
func (sem *Semaphore) Block() {
	sem.rw.Lock()
	sem.wg.Wait()
}

// Unblock allows the semaphore to be locked again.
func (sem *Semaphore) Unblock() {
	sem.rw.Unlock()
}

// Do executes the given function synchronously.
func (sem *Semaphore) Do(fn func()) {
	sem.Lock()
	sem.wg.Add(1)

	defer sem.Unlock()
	defer sem.wg.Done()

	fn()
}

func (sem *Semaphore) SetPanicHandler(panicHandler PanicHandler) {
	sem.panicHandler = panicHandler
}

func (sem *Semaphore) handlePanic() {
	if sem.panicHandler != nil {
		sem.panicHandler.HandlePanic()
	}
}

// Go executes the given function asynchronously.
func (sem *Semaphore) Go(fn func()) {
	defer sem.handlePanic()

	sem.Lock()
	sem.wg.Add(1)

	go func() {
		defer sem.Unlock()
		defer sem.wg.Done()

		fn()
	}()
}

// Wait waits for all functions started by Go to finish executing.
func (sem *Semaphore) Wait() {
	sem.wg.Wait()
}
