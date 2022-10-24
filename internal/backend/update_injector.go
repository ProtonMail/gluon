package backend

import (
	"context"
	"sync"

	"github.com/ProtonMail/gluon/connector"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/logging"
)

// updateInjector allows anyone to publish custom imap updates alongside the updates that are generated from the
// Connector.
type updateInjector struct {
	// updatesCh is the channel that delivers API updates to the mailserver.
	updatesCh chan imap.Update

	// forwardWG is used to ensure we wait until the forward() goroutine has finished executing.
	forwardWG     sync.WaitGroup
	forwardQuitCh chan struct{}
}

// newUpdateInjector creates a new updateInjector.
//
// nolint:contextcheck
func newUpdateInjector(connector connector.Connector, userID string) *updateInjector {
	injector := &updateInjector{
		updatesCh:     make(chan imap.Update),
		forwardQuitCh: make(chan struct{}),
	}

	injector.forwardWG.Add(1)

	logging.GoAnnotated(context.Background(), func(ctx context.Context) {
		injector.forward(ctx, connector.GetUpdates())
	}, logging.Labels{
		"Action": "Forwarding updates",
		"UserID": userID,
	})

	return injector
}

// GetUpdates returns a channel on which updates from the server are sent.
func (u *updateInjector) GetUpdates() <-chan imap.Update {
	return u.updatesCh
}

func (u *updateInjector) Close(ctx context.Context) error {
	close(u.forwardQuitCh)
	u.forwardWG.Wait()

	return nil
}

// forward pulls updates off the stream and forwards them to the outgoing update channel.
func (u *updateInjector) forward(ctx context.Context, updateCh <-chan imap.Update) {
	defer func() {
		close(u.updatesCh)
		u.forwardWG.Done()
	}()

	for {
		select {
		case update, ok := <-updateCh:
			if !ok {
				return
			}

			u.send(ctx, update)

		case <-u.forwardQuitCh:
			return
		}
	}
}

// send the update on the updates channel, optionally blocking until it has been processed.
func (u *updateInjector) send(ctx context.Context, update imap.Update, withBlock ...bool) {
	select {
	case <-u.forwardQuitCh:
		return

	case u.updatesCh <- update:
		if len(withBlock) > 0 && withBlock[0] {
			update.WaitContext(ctx)
		}
	}
}
