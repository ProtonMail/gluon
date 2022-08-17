package store_benchmarks

import (
	"context"
	"github.com/ProtonMail/gluon/imap"
	"math/rand"

	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/benchmark"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/flags"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/reporter"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/timing"
	"github.com/ProtonMail/gluon/store"
)

type Get struct {
	uuids []imap.InternalMessageID
}

func (*Get) Name() string {
	return "store-get"
}

func (g *Get) Setup(ctx context.Context, s store.Store) error {
	uuids, err := CreateRandomState(s, *flags.StoreItemCount)
	if err != nil {
		return err
	}

	g.uuids = uuids

	return nil
}

func (*Get) TearDown(ctx context.Context, store store.Store) error {
	return nil
}

func (g *Get) Run(ctx context.Context, st store.Store) (*reporter.BenchmarkRun, error) {
	uuidLen := len(g.uuids)

	return RunStoreWorkers(ctx, st, func(ctx context.Context, s store.Store, dc *timing.Collector, u uint) error {
		for i := 0; i < uuidLen; i++ {
			index := rand.Intn(uuidLen)

			dc.Start()
			_, err := s.Get(g.uuids[index])
			dc.Stop()

			if err != nil {
				panic(err)
			}

		}

		return nil
	}), nil
}

func init() {
	benchmark.RegisterBenchmark(NewStoreBenchmarkRunner(&Get{}))
}
