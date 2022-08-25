package remote

import (
	"context"
	"runtime/pprof"
	"sync"

	"github.com/ProtonMail/gluon/connector"
	"github.com/ProtonMail/gluon/imap"
)

// User performs operations against a remote server using a connector.
type User struct {
	userID string

	// conn is what the user uses to perform API operations.
	conn connector.Connector

	// updatesCh is the channel that delivers API updates to the mailserver.
	updatesCh chan imap.Update

	// lastOp holds an operation while it has been popped off the queue but not yet executed.
	connMetadataStore     connMetadataStore
	connMetadataStoreLock sync.RWMutex

	// forwardWG is used to ensure we wait until the forward() goroutine has finished executing.
	forwardWG     sync.WaitGroup
	forwardQuitCh chan struct{}
}

// newUser constructs a new user with the given (IMAP) credentials.
// It serializes its operation queue to a file at the given filepath,
// and performs remote operations using the given connector.
func newUser(ctx context.Context, userID string, conn connector.Connector) (*User, error) {
	user := &User{
		userID:            userID,
		conn:              conn,
		updatesCh:         make(chan imap.Update),
		connMetadataStore: newConnMetadataStore(),
		forwardQuitCh:     make(chan struct{}),
	}

	// send connector updates along to the mailserver.
	user.forwardWG.Add(1)

	go func() {
		labels := pprof.Labels("go", "forward()", "UserID", userID)
		pprof.Do(ctx, labels, func(_ context.Context) {
			user.forward(conn.GetUpdates())
		})
	}()

	return user, nil
}

// GetUpdates returns a channel on which updates from the server are sent.
func (user *User) GetUpdates() <-chan imap.Update {
	return user.updatesCh
}

func (user *User) Close(ctx context.Context) error {
	close(user.forwardQuitCh)
	user.forwardWG.Wait()

	if err := user.conn.Close(ctx); err != nil {
		return err
	}

	return nil
}

// forward pulls updates off the stream and forwards them to the outgoing update channel.
func (user *User) forward(updateCh <-chan imap.Update) {
	defer func() {
		close(user.updatesCh)
		user.forwardWG.Done()
	}()

	for {
		select {
		case update := <-updateCh:
			user.send(update)

		case <-user.forwardQuitCh:
			return
		}
	}
}

// send sends the update on the user's updates channel, optionally blocking until it has been processed.
func (user *User) send(update imap.Update, withBlock ...bool) {
	select {
	case user.updatesCh <- update:

	case <-user.forwardQuitCh:
		return
	}

	if len(withBlock) > 0 && withBlock[0] {
		update.Wait()
	}
}
