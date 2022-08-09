package store_benchmarks

import (
	"context"

	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/benchmark"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/flags"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/reporter"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/utils"
	"github.com/ProtonMail/gluon/store"
	"github.com/google/uuid"
)

type Rename struct {
	uuids        []string
	uuidsRenamed []string
}

func (*Rename) Name() string {
	return "store-rename"
}

func (r *Rename) Setup(ctx context.Context, s store.Store) error {
	uuids, err := CreateRandomState(s, *flags.StoreItemCount)
	if err != nil {
		return err
	}

	r.uuids = uuids
	r.uuidsRenamed = make([]string, len(r.uuids))

	for i := 0; i < len(r.uuids); i++ {
		r.uuidsRenamed[i] = uuid.NewString()
	}

	return nil
}

func (*Rename) TearDown(ctx context.Context, store store.Store) error {
	return nil
}

func (r *Rename) Run(ctx context.Context, st store.Store) (*reporter.BenchmarkRun, error) {
	return RunStoreWorkersSplitRange(ctx, st, uint(len(r.uuids)), func(ctx context.Context, s store.Store, dc *utils.DurationCollector, start, end uint) error {
		for i := start; i < end; i++ {
			dc.Start()
			err := s.Update(r.uuids[i], r.uuidsRenamed[i])
			dc.Stop()

			if err != nil {
				panic(err)
			}
		}

		return nil
	}), nil
}

func init() {
	benchmark.RegisterBenchmark(NewStoreBenchmarkRunner(&Rename{}))
}
