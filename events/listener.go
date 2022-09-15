package events

import "net"

type EventListenerAdded struct {
	eventBase

	Addr net.Addr
}

type EventListenerRemoved struct {
	eventBase

	Addr net.Addr
}
