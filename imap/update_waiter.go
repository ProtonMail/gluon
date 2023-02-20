package imap

import (
	"context"
)

type Waiter interface {
	// Wait waits until the update has been marked as done.
	Wait() (error, bool)

	// WaitContext waits until the update has been marked as done or the context is cancelled.
	WaitContext(context.Context) (error, bool)

	// Done marks the update as done and report an error (if any).
	Done(error)
}

type updateWaiter struct {
	waitCh chan error
}

func newUpdateWaiter() *updateWaiter {
	return &updateWaiter{
		waitCh: make(chan error, 1),
	}
}

func (w *updateWaiter) Wait() (error, bool) {
	err, ok := <-w.waitCh
	return err, ok
}

func (w *updateWaiter) WaitContext(ctx context.Context) (error, bool) {
	select {
	case <-ctx.Done():
		return nil, false
	case err, ok := <-w.waitCh:
		return err, ok
	}
}

func (w *updateWaiter) Done(err error) {
	if err != nil {
		w.waitCh <- err
	}

	close(w.waitCh)
}
