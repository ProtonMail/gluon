package store_benchmarks

import (
	"context"
	"math/rand"

	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/benchmark"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/flags"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/reporter"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/timing"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/utils"
	"github.com/ProtonMail/gluon/store"
	"github.com/google/uuid"
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
		messages := []string{utils.MessageAfterNoonMeeting, utils.MessageMultiPartMixed, utils.MessageEmbedded}
		messagesLen := len(messages)

		for i := uint(0); i < *flags.StoreItemCount; i++ {
			dc.Start()
			err := s.Set(uuid.NewString(), []byte(messages[rand.Intn(messagesLen)]))
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
