package tests

import (
	"sync"

	"github.com/ProtonMail/gluon"
	"github.com/ProtonMail/gluon/events"
)

type eventWaiter struct {
	wg     sync.WaitGroup
	ch     <-chan events.Event
	server *gluon.Server
}

func newEventWaiter(server *gluon.Server) *eventWaiter {
	return &eventWaiter{
		wg:     sync.WaitGroup{},
		ch:     server.AddWatcher(),
		server: server,
	}
}

func (e *eventWaiter) waitEndOfSession() {
	e.wg.Add(1)

	go func() {
		for message := range e.ch {
			switch message.(type) {
			case events.EventSessionRemoved:
				e.wg.Done()
				return
			}
		}
	}()

	e.wg.Wait()
}
