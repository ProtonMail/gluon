package imap

import (
	"context"
	"sync"
)

type Waiter interface {
	// Wait waits until the update has been marked as done.
	Wait()

	// WaitContext waits until the update has been marked as done or the context is cancelled.
	WaitContext(context.Context)

	// Done marks the update as done.
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

func (w *updateWaiter) WaitContext(ctx context.Context) {
	waitCh := make(chan struct{})

	go func() { w.wg.Wait(); close(waitCh) }()

	select {
	case <-ctx.Done():
	case <-waitCh:
	}
}

func (w *updateWaiter) Done() {
	w.wg.Done()
}
