package remote

import (
	"errors"
	"io"
	"os"
	"sync"

	"github.com/ProtonMail/gluon/connector"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/pchan"
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

	// queue is channel of operations that must be performed on the API.
	queue *pchan.PChan[operation]

	// lastOp holds an operation while it has been popped off the queue but not yet executed.
	lastOp operation

	// closed holds whether the operation queue has been closed.
	closed     bool
	closedLock sync.RWMutex
}

// newUser constructs a new user with the given (IMAP) credentials.
// It serializes its operation queue to a file at the given filepath,
// and performs remote operations using the given connector.
func newUser(userID, path string, conn connector.Connector) (*User, error) {
	user := &User{
		userID:    userID,
		path:      path,
		conn:      conn,
		updatesCh: make(chan imap.Update),
		queue:     pchan.New[operation](),
	}

	// load any saved operations that were not processed fully before.
	if err := user.load(); err != nil {
		return nil, err
	}

	// send connector updates along to the mailserver.
	go user.forward(conn.GetUpdates())

	// process remote operations on the operation queue.
	go user.process()

	return user, nil
}

// GetUpdates returns a channel on which updates from the server are sent.
func (user *User) GetUpdates() <-chan imap.Update {
	return user.updatesCh
}

// Close closes the remote user.
func (user *User) Close() error {
	ops, err := user.closeQueue()
	if err != nil {
		return err
	}

	if user.lastOp != nil {
		ops = append([]operation{user.lastOp}, ops...)
	}

	b, err := saveOps(ops)
	if err != nil {
		return err
	}

	if err := os.WriteFile(user.path, b, 0o600); err != nil {
		return err
	}

	if err := user.conn.Close(); err != nil {
		return err
	}

	return nil
}

// forward pulls updates off the stream and forwards them to the outgoing update channel.
func (user *User) forward(updateCh <-chan imap.Update) {
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

// load reads queued remote operations from disk and fills the operation queue with them.
func (user *User) load() error {
	f, err := os.Open(user.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}

	b, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	ops, err := loadOps(b)
	if err != nil {
		return err
	}

	if err := os.Remove(user.path); err != nil {
		return err
	}

	for _, op := range ops {
		if err := user.pushOp(op); err != nil {
			return err
		}
	}

	return nil
}
