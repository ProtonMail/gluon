package server

import (
	"context"
	"net"

	"github.com/ProtonMail/gluon/profiling"
)

type Server interface {
	// Close should close all server connections or shut down the server.
	Close(ctx context.Context) error

	// Address should return the server address.
	Address() net.Addr
}

type ServerBuilder interface {
	// New Create new Server instance at a given path and use the command profiler, if possible.
	New(ctx context.Context, serverPath string, profiler profiling.CmdProfilerBuilder) (Server, error)
}
