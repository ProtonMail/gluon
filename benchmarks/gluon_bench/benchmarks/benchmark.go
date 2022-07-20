package benchmarks

import (
	"context"
	"net"
)

type Benchmark interface {
	// Name should return the name of the bechmark, it will also be used to match against cli args.
	Name() string

	// Setup sets up the benchmark state, this is not timed.
	Setup(ctx context.Context, addr net.Addr) error

	// Run performs the actual benchmark, this is timed.
	Run(ctx context.Context, addr net.Addr) error

	// TearDown clear the benchmark state, this is not timed.
	TearDown(ctx context.Context, addr net.Addr) error
}
