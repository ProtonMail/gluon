package remote

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"
)

var ErrQueueClosed = errors.New("the queue is closed")

// process repeatedly pulls items off the operation queue and executes them.
// TODO: What should we do with operations that failed to execute due to auth reasons?
// We might want to save them somewhere so we can try again after the user has logged back in.
func (user *User) process() {
	defer user.processWG.Done()

	for {
		op, ok := user.popOp()
		if !ok {
			return
		}

		user.lastOp = op

		if err := user.execute(context.Background(), op); err != nil {
			logrus.WithField("op", op).WithError(err).Error("Error handling remote operation")
		}

		user.lastOp = nil
	}
}

// pushOp enqueues the given remote operation and pauses the remote update stream.
func (user *User) pushOp(op operation) error {
	user.closedLock.RLock()
	defer user.closedLock.RUnlock()

	if user.closed {
		return ErrQueueClosed
	}

	user.conn.Pause()

	user.queue.Push(op)

	return nil
}

// popOp pops the next remote operation off the queue and, if the queue is empty, resumes the update stream.
func (user *User) popOp() (operation, bool) {
	op, ok := user.queue.Pop()
	if !ok {
		return nil, false
	}

	for {
		next, ok := user.queue.Peek()
		if !ok {
			break
		}

		merged, ok := op.merge(next)
		if !ok {
			break
		}

		if _, ok := user.queue.Pop(); !ok {
			panic("the queue should not be empty")
		}

		op = merged
	}

	if user.queue.Len() == 0 {
		user.conn.Resume()
	}

	return op, true
}

func (user *User) setMailboxID(tempID, mboxID string) {
	user.queue.Apply(func(op operation) {
		switch op := op.(type) {
		case mailboxOperation:
			op.setMailboxID(tempID, mboxID)
		}
	})
}

func (user *User) setMessageID(tempID, messageID string) {
	user.queue.Apply(func(op operation) {
		switch op := op.(type) {
		case messageOperation:
			op.setMessageID(tempID, messageID)
		}
	})
}

func (user *User) closeQueue() ([]operation, error) {
	user.closedLock.Lock()
	defer user.closedLock.Unlock()

	if user.closed {
		panic("the queue is already closed")
	}

	user.closed = true

	return user.queue.Close(), nil
}
