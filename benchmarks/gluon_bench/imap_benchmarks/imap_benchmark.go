package imap_benchmarks

import (
	"context"
	"net"
)

// IMAPBenchmark is intended to be used to build benchmarks which bench IMAP commands on a given server.
type IMAPBenchmark interface {
	// Name should return the name of the benchmark. It will also be used to match against cli args.
	Name() string

	// Setup sets up the benchmark state, this is not timed.
	Setup(ctx context.Context, addr net.Addr) error

	// Run performs the actual benchmark, this is timed.
	Run(ctx context.Context, addr net.Addr) error

	// TearDown clear the benchmark state, this is not timed.
	TearDown(ctx context.Context, addr net.Addr) error
}
