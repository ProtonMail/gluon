package imap_benchmarks

import (
	"context"
	"fmt"

	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/benchmark"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/flags"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/imap_benchmarks/server"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/reporter"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/utils"
)

type IMAPBenchmarkRunner struct {
	benchmark          IMAPBenchmark
	serverBuilder      server.ServerBuilder
	cmdProfilerBuilder *utils.DurationCmdProfilerBuilder
	server             server.Server
}

func (i *IMAPBenchmarkRunner) Name() string {
	return i.benchmark.Name()
}

// Setup sets up the benchmark state, this is not timed.
func (i *IMAPBenchmarkRunner) Setup(ctx context.Context, benchmarkDir string) error {
	i.cmdProfilerBuilder.Clear()

	server, err := i.serverBuilder.New(ctx, benchmarkDir, i.cmdProfilerBuilder)
	if err != nil {
		return err
	}

	i.server = server

	if err := i.benchmark.Setup(ctx, i.server.Address()); err != nil {
		return err
	}

	return nil
}

// Run performs the actual benchmark, this is timed.
func (i *IMAPBenchmarkRunner) Run(ctx context.Context) (*reporter.BenchmarkRun, error) {
	scopedTimer := utils.ScopedTimer{}

	scopedTimer.Start()

	err := i.benchmark.Run(ctx, i.server.Address())

	scopedTimer.Stop()

	if err != nil {
		return nil, err
	}

	return NewIMAPBenchmarkRun(scopedTimer.Elapsed(), i.cmdProfilerBuilder.Merge()), nil
}

// TearDown clear the benchmark state, this is not timed.
func (i *IMAPBenchmarkRunner) TearDown(ctx context.Context) error {
	if i.server != nil {
		if err := i.benchmark.TearDown(ctx, i.server.Address()); err != nil {
			return err
		}

		if err := i.server.Close(ctx); err != nil {
			return err
		}
	}

	return nil
}

func NewIMAPBenchmarkRunner(bench IMAPBenchmark) benchmark.Benchmark {
	var serverBuilder server.ServerBuilder

	if len(*flags.IMAPRemoteServer) != 0 {
		builder, err := server.NewRemoteServerBuilder(*flags.IMAPRemoteServer)
		if err != nil {
			panic(fmt.Sprintf("Invalid Server address: %v", err))
		}

		serverBuilder = builder
	} else {
		serverBuilder = &server.LocalServerBuilder{}
	}

	return &IMAPBenchmarkRunner{benchmark: bench,
		serverBuilder:      serverBuilder,
		cmdProfilerBuilder: utils.NewDurationCmdProfilerBuilder(),
	}
}
