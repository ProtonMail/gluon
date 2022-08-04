package benchmark

import (
	"context"

	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/reporter"
)

type Benchmark interface {
	// Name should return the name of the benchmark. It will also be used to match against cli args.
	Name() string

	// Setup sets up the benchmark state, this is not timed.
	Setup(ctx context.Context, benchmarkDir string) error

	// Run performs the actual benchmark, this is timed.
	Run(ctx context.Context) (*reporter.BenchmarkRun, error)

	// TearDown clear the benchmark state, this is not timed.
	TearDown(ctx context.Context) error
}
