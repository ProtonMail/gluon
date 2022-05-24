package imap

import "sync"

type Waiter interface {
	Wait()
	Done()
}

type updateWaiter struct {
	wg sync.WaitGroup
}

func newUpdateWaiter() *updateWaiter {
	var result updateWaiter

	result.wg.Add(1)

	return &result
}

func (w *updateWaiter) Wait() {
	w.wg.Wait()
}

func (w *updateWaiter) Done() {
	w.wg.Done()
}
