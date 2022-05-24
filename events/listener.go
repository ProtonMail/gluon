package events

import "net"

type EventListenerAdded struct {
	Addr net.Addr
}

func (EventListenerAdded) _isEvent() {}

type EventListenerRemoved struct {
	Addr net.Addr
}

func (EventListenerRemoved) _isEvent() {}
