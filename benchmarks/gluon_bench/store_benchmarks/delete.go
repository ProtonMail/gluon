package store_benchmarks

import (
	"context"
	"github.com/ProtonMail/gluon/imap"

	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/benchmark"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/flags"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/reporter"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/timing"
	"github.com/ProtonMail/gluon/store"
)

type Delete struct {
	uuids []imap.InternalMessageID
}

func (*Delete) Name() string {
	return "store-delete"
}

func (d *Delete) Setup(ctx context.Context, s store.Store) error {
	uuids, err := CreateRandomState(s, *flags.StoreItemCount)
	if err != nil {
		return err
	}

	d.uuids = uuids

	return nil
}

func (*Delete) TearDown(ctx context.Context, store store.Store) error {
	return nil
}

func (d *Delete) Run(ctx context.Context, st store.Store) (*reporter.BenchmarkRun, error) {
	return RunStoreWorkersSplitRange(ctx, st, uint(len(d.uuids)), func(ctx context.Context, s store.Store, dc *timing.Collector, start, end uint) error {
		for i := start; i < end; i++ {
			dc.Start()
			err := s.Delete(d.uuids[i])
			dc.Stop()

			if err != nil {
				panic(err)
			}
		}

		return nil
	}), nil
}

func init() {
	benchmark.RegisterBenchmark(NewStoreBenchmarkRunner(&Delete{}))
}
