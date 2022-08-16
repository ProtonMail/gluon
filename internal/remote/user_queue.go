package remote

import (
	"context"
	"errors"

	"github.com/bradenaw/juniper/xslices"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/queue"
	"github.com/sirupsen/logrus"
)

var ErrQueueClosed = errors.New("the queue is closed")

type userOpQueue struct {
	poppedOps          []operation
	queue              *queue.CTQueue[operation]
	tempMailboxIDTable map[string]string
	tempMessageIDTable map[string]string
}

func newUserOpQueue() userOpQueue {
	return userOpQueue{
		queue:              queue.NewCTQueue[operation](),
		tempMessageIDTable: make(map[string]string),
		tempMailboxIDTable: make(map[string]string),
	}
}

func (uoq *userOpQueue) popAndMerge() operation {
	var firstOp operation

	if firstOp = uoq.getNextOp(); firstOp == nil {
		return nil
	}

	for {
		var secondOp operation
		if secondOp = uoq.tryGetNextOp(); secondOp == nil {
			return firstOp
		}

		if mergedOp, ok := firstOp.merge(secondOp); ok {
			firstOp = mergedOp
			continue
		}

		uoq.poppedOps = append(uoq.poppedOps, secondOp)

		return firstOp
	}
}

func (uoq *userOpQueue) getNextOp() operation {
	if len(uoq.poppedOps) != 0 {
		item := uoq.poppedOps[0]
		uoq.poppedOps = uoq.poppedOps[1:]

		return item
	}

	if item, ok := uoq.queue.Pop(); ok {
		return item
	}

	return nil
}

func (uoq *userOpQueue) tryGetNextOp() operation {
	if len(uoq.poppedOps) != 0 {
		item := uoq.poppedOps[0]
		uoq.poppedOps = uoq.poppedOps[1:]

		return item
	}

	if item, ok := uoq.queue.TryPop(); ok {
		return item
	}

	return nil
}

func (user *User) newContextWithIMAPID(ctx context.Context, id ConnMetadataID) context.Context {
	if v := user.connMetadataStore.GetValue(id, imap.IMAPIDConnMetadataKey); v != nil {
		switch x := v.(type) {
		case imap.ID:
			ctx = imap.NewContextWithIMAPID(ctx, x)
		}
	}

	return ctx
}

// process repeatedly pulls items off the operation queue and executes them.
// TODO: What should we do with operations that failed to execute due to auth reasons?
// We might want to save them somewhere so we can try again after the user has logged back in.
func (user *User) process() {
	defer user.processWG.Done()

	for {
		// Pops the next remote operation off the queue.
		op := user.opQueue.popAndMerge()

		if op == nil {
			return
		}
		// if the queue is empty, resumes the update stream
		if user.opQueue.queue.Len() == 0 {
			user.conn.Resume()
		}

		user.lastOp = op
		if err := user.execute(user.newContextWithIMAPID(context.Background(), op.getConnMetadataID()), op); err != nil {
			logrus.WithField("op", op).WithError(err).Error("Error handling remote operation")
		}

		user.lastOp = nil
	}
}

// pushOp enqueues the given remote operation and pauses the remote update stream.
func (uoq *userOpQueue) pushOp(user *User, op operation) error {
	if uoq.queue.IsClosed() {
		return ErrQueueClosed
	}

	user.conn.Pause()

	uoq.queue.Push(op)

	return nil
}

func (uoq *userOpQueue) closeQueueAndRetrieveRemaining() ([]operation, error) {
	return uoq.queue.CloseAndRetrieveRemaining(), nil
}

func (uoq *userOpQueue) closeQueue() {
	uoq.queue.Close()
}

func (user *User) pushOp(op operation) error {
	return user.opQueue.pushOp(user, op)
}

func (uoq *userOpQueue) addMailboxTempID(tmpID string, realID string) {
	uoq.tempMailboxIDTable[tmpID] = realID
}

func (uoq *userOpQueue) addMessageTempID(tmpID string, realID string) {
	uoq.tempMessageIDTable[tmpID] = realID
}

func (uoq *userOpQueue) remMessageTempID(tmpID string) {
	delete(uoq.tempMessageIDTable, tmpID)
}

func (uoq *userOpQueue) remMailboxTempID(tmpID string) {
	delete(uoq.tempMailboxIDTable, tmpID)
}

func (uoq *userOpQueue) translateMailboxIDs(ids ...string) []string {
	if len(uoq.tempMailboxIDTable) == 0 {
		return ids
	}

	return xslices.Map(ids, func(id string) string {
		if v, ok := uoq.tempMailboxIDTable[id]; ok {
			return v
		}

		return id
	})
}

func (uoq *userOpQueue) translateMessageIDs(ids ...string) []string {
	if len(uoq.tempMessageIDTable) == 0 {
		return ids
	}

	return xslices.Map(ids, func(id string) string {
		if v, ok := uoq.tempMessageIDTable[id]; ok {
			return v
		}

		return id
	})
}

func (uoq *userOpQueue) translateMailboxID(id string) string {
	return uoq.translateMailboxIDs(id)[0]
}

func (uoq *userOpQueue) translateMessageID(id string) string {
	return uoq.translateMessageIDs(id)[0]
}
