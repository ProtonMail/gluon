package events

import "net"

type SessionAdded struct {
	eventBase

	SessionID  int
	LocalAddr  net.Addr
	RemoteAddr net.Addr
}

type SessionRemoved struct {
	eventBase

	SessionID int
}
