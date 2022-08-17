package store_benchmarks

import (
	"context"
	"github.com/ProtonMail/gluon/imap"
	"sync"
	"time"

	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/flags"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/reporter"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/timing"
	"github.com/ProtonMail/gluon/store"
	"github.com/google/uuid"
)

func CreateRandomState(store store.Store, count uint) ([]imap.InternalMessageID, error) {
	uuids := make([]imap.InternalMessageID, 0, count)
	data := make([]byte, *flags.StoreItemSize)

	for i := uint(0); i < count; i++ {
		uuid := imap.InternalMessageID(uuid.NewString())

		if err := store.Set(uuid, data); err != nil {
			return nil, err
		}

		uuids = append(uuids, uuid)
	}

	return uuids, nil
}

func RunStoreWorkers(ctx context.Context, st store.Store, fn func(context.Context, store.Store, *timing.Collector, uint) error) *reporter.BenchmarkRun {
	wg := sync.WaitGroup{}

	durations := make([]time.Duration, 0, *flags.StoreWorkers**flags.StoreItemCount)
	collectors := make([]*timing.Collector, *flags.StoreWorkers)

	for i := uint(0); i < *flags.StoreWorkers; i++ {
		wg.Add(1)

		go func(index uint) {
			defer wg.Done()

			collector := timing.NewDurationCollector(int(*flags.StoreItemCount))

			if err := fn(ctx, st, collector, index); err != nil {
				panic(err)
			}

			collectors[index] = collector
		}(i)
	}

	wg.Wait()

	for _, v := range collectors {
		durations = append(durations, v.Durations()...)
	}

	return reporter.NewBenchmarkRun(durations, nil)
}

func RunStoreWorkersSplitRange(ctx context.Context, st store.Store, length uint, fn func(context.Context, store.Store, *timing.Collector, uint, uint) error) *reporter.BenchmarkRun {
	workDivision := length / *flags.StoreWorkers

	return RunStoreWorkers(ctx, st, func(ctx context.Context, s store.Store, collector *timing.Collector, u uint) error {
		end := workDivision * (u + 1)
		if end > length {
			end = length
		}

		return fn(ctx, st, collector, u*workDivision, end)
	})
}
