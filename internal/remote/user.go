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

	// path is the path at which the operation queue will be saved to disk.
	path string

	// conn is what the user uses to perform API operations.
	conn connector.Connector

	// updatesCh is the channel that delivers API updates to the mailserver.
	updatesCh chan imap.Update

	// lastOp holds an operation while it has been popped off the queue but not yet executed.
	connMetadataStore     connMetadataStore
	connMetadataStoreLock sync.RWMutex

	// forwardWG is used to ensure we wait until the forward() goroutine has finished executing.
	forwardWG sync.WaitGroup
}

// newUser constructs a new user with the given (IMAP) credentials.
// It serializes its operation queue to a file at the given filepath,
// and performs remote operations using the given connector.
func newUser(ctx context.Context, userID, path string, conn connector.Connector) (*User, error) {
	user := &User{
		userID:            userID,
		path:              path,
		conn:              conn,
		updatesCh:         make(chan imap.Update),
		connMetadataStore: newConnMetadataStore(),
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
	if err := user.conn.Close(ctx); err != nil {
		return err
	}

	//TODO: GODT-1647 fix double call to Close().
	if user.updatesCh != nil {
		user.forwardWG.Wait()
		user.updatesCh = nil
	}

	return nil
}

// forward pulls updates off the stream and forwards them to the outgoing update channel.
func (user *User) forward(updateCh <-chan imap.Update) {
	defer func() {
		close(user.updatesCh)
		user.forwardWG.Done()
	}()

	for update := range updateCh {
		user.send(update)
	}
}

// send sends the update on the user's updates channel, optionally blocking until it has been processed.
func (user *User) send(update imap.Update, withBlock ...bool) {
	user.updatesCh <- update

	if len(withBlock) > 0 && withBlock[0] {
		update.Wait()
	}
}
