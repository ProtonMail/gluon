package imap

import (
	"context"
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
	waitCh chan struct{}
}

func newUpdateWaiter() *updateWaiter {
	return &updateWaiter{
		waitCh: make(chan struct{}),
	}
}

func (w *updateWaiter) Wait() {
	<-w.waitCh
}

func (w *updateWaiter) WaitContext(ctx context.Context) {
	select {
	case <-ctx.Done():
	case <-w.waitCh:
	}
}

func (w *updateWaiter) Done() {
	close(w.waitCh)
}
