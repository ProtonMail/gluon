package store_benchmarks

import (
	"context"
	"os"

	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/benchmark"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/flags"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/reporter"
	"github.com/ProtonMail/gluon/store"
)

type StoreBenchmark interface {
	// Name returns benchmark's name.
	Name() string

	// Setup should prepare the benchmark.
	Setup(ctx context.Context, store store.Store) error

	// TearDown should clean the benchmark.
	TearDown(ctx context.Context, store store.Store) error

	// Run the benchmark.
	Run(ctx context.Context, store store.Store) (*reporter.BenchmarkRun, error)
}

type StoreBenchmarkRunner struct {
	benchmark    StoreBenchmark
	benchmarkDir string
	store        store.Store
}

func (s *StoreBenchmarkRunner) Name() string {
	return s.benchmark.Name()
}

func (s *StoreBenchmarkRunner) Setup(ctx context.Context, benchmarkDir string) error {
	store, err := NewStore(*flags.Store, benchmarkDir)
	if err != nil {
		return err
	}

	s.store = store
	s.benchmarkDir = benchmarkDir

	if err := s.benchmark.Setup(ctx, s.store); err != nil {
		return err
	}

	return nil
}

func (s *StoreBenchmarkRunner) Run(ctx context.Context) (*reporter.BenchmarkRun, error) {
	benchRuns, err := s.benchmark.Run(ctx, s.store)
	if err != nil {
		return nil, err
	}

	return benchRuns, nil
}

func (s *StoreBenchmarkRunner) TearDown(ctx context.Context) error {
	if err := s.benchmark.TearDown(ctx, s.store); err != nil {
		return err
	}

	if err := s.store.Close(); err != nil {
		return err
	}

	if err := os.RemoveAll(s.benchmarkDir); err != nil {
		return err
	}

	return nil
}

func NewStoreBenchmarkRunner(bench StoreBenchmark) benchmark.Benchmark {
	return &StoreBenchmarkRunner{benchmark: bench}
}
