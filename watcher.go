package gluon

import (
	"reflect"

	"github.com/ProtonMail/gluon/events"
	"github.com/ProtonMail/gluon/internal/queue"
)

type watcher struct {
	types   map[reflect.Type]struct{}
	eventCh *queue.QueuedChannel[events.Event]
}

func newWatcher(ofType ...events.Event) *watcher {
	types := make(map[reflect.Type]struct{}, len(ofType))

	for _, t := range ofType {
		types[reflect.TypeOf(t)] = struct{}{}
	}

	return &watcher{
		types:   types,
		eventCh: queue.NewQueuedChannel[events.Event](1, 1),
	}
}

func (w *watcher) isWatching(event events.Event) bool {
	if len(w.types) == 0 {
		return true
	}

	_, ok := w.types[reflect.TypeOf(event)]
	return ok
}

func (w *watcher) getChannel() <-chan events.Event {
	return w.eventCh.GetChannel()
}

func (w *watcher) send(event events.Event) bool {
	return w.eventCh.Enqueue(event)
}

func (w *watcher) close() {
	w.eventCh.Close()
}
