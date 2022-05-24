package events

import "net"

type EventSessionAdded struct {
	SessionID  int
	LocalAddr  net.Addr
	RemoteAddr net.Addr
}

func (EventSessionAdded) _isEvent() {}

type EventSessionRemoved struct {
	SessionID int
}

func (EventSessionRemoved) _isEvent() {}
