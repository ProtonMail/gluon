package events

import "net"

type ListenerAdded struct {
	eventBase

	Addr net.Addr
}

type ListenerRemoved struct {
	eventBase

	Addr net.Addr
}
