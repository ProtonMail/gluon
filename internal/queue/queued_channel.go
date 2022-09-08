package queue

// QueuedChannel represents a channel on which queued items can be published without having to worry if the reader
// has actually consumed existing items first or if there's no way of knowing ahead of time what the ideal channel
// buffer size should be.
type QueuedChannel[T any] struct {
	ch    chan T
	queue *CTQueue[T]
}

func NewQueuedChannel[T any](channelBufferSize int, capacity int) *QueuedChannel[T] {
	queue := &QueuedChannel[T]{
		ch:    make(chan T, channelBufferSize),
		queue: NewCTQueueWithCapacity[T](capacity),
	}

	go func() {
		defer close(queue.ch)

		for {
			item, ok := queue.queue.Pop()
			if !ok {
				return
			}

			queue.ch <- item
		}
	}()

	return queue
}

func (q *QueuedChannel[T]) Queue(items ...T) bool {
	return q.queue.PushMany(items...)
}

func (q *QueuedChannel[T]) GetChannel() <-chan T {
	return q.ch
}

func (q *QueuedChannel[T]) Close() {
	q.queue.Close()
}
