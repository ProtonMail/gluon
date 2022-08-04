package server

import (
	"context"
	"net"

	"github.com/ProtonMail/gluon/profiling"
)

// RemoteServer can't control the start or stopping of the server but can still be used to run the benchmarks
// against an existing server.
type RemoteServer struct {
	address net.Addr
}

func (*RemoteServer) Close(ctx context.Context) error {
	return nil
}

func (r *RemoteServer) Address() net.Addr {
	return r.address
}

type RemoteServerBuilder struct {
	address net.Addr
}

func NewRemoteServerBuilder(address string) (*RemoteServerBuilder, error) {
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}

	return &RemoteServerBuilder{address: addr}, nil
}

func (r *RemoteServerBuilder) New(ctx context.Context, serverPath string, profiler profiling.CmdProfilerBuilder) (Server, error) {
	return &RemoteServer{address: r.address}, nil
}
