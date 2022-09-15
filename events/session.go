package events

import "net"

type EventSessionAdded struct {
	eventBase

	SessionID  int
	LocalAddr  net.Addr
	RemoteAddr net.Addr
}

type EventSessionRemoved struct {
	eventBase

	SessionID int
}
