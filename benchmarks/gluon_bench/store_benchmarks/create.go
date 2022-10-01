package store_benchmarks

import (
	"context"

	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/benchmark"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/flags"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/reporter"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/timing"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/store"
)

type Create struct{}

func (*Create) Name() string {
	return "store-create"
}

func (*Create) Setup(ctx context.Context, store store.Store) error {
	return nil
}

func (*Create) TearDown(ctx context.Context, store store.Store) error {
	return nil
}

func (*Create) Run(ctx context.Context, st store.Store) (*reporter.BenchmarkRun, error) {
	return RunStoreWorkers(ctx, st, func(ctx context.Context, s store.Store, dc *timing.Collector, u uint) error {
		data := make([]byte, *flags.StoreItemSize)

		for i := uint(0); i < *flags.StoreItemCount; i++ {
			dc.Start()
			err := s.Set(imap.InternalMessageID(uint64(i)), data)
			dc.Stop()

			if err != nil {
				return err
			}

		}

		return nil
	}), nil
}

func init() {
	benchmark.RegisterBenchmark(NewStoreBenchmarkRunner(&Create{}))
}
