package store_benchmarks

import (
	"context"
	"sync"
	"time"

	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/flags"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/reporter"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/utils"
	"github.com/ProtonMail/gluon/store"
	"github.com/google/uuid"
)

func CreateRandomState(store store.Store, count uint) ([]string, error) {
	uuids := make([]string, 0, count)
	data := make([]byte, *flags.StoreItemSize)

	for i := uint(0); i < count; i++ {
		uuid := uuid.NewString()
		if err := store.Set(uuid, data); err != nil {
			return nil, nil
		}

		uuids = append(uuids, uuid)
	}

	return uuids, nil
}

func RunStoreWorkers(ctx context.Context, st store.Store, fn func(context.Context, store.Store, *utils.DurationCollector, uint) error) *reporter.BenchmarkRun {
	wg := sync.WaitGroup{}

	durations := make([]time.Duration, 0, *flags.StoreWorkers**flags.StoreItemCount)
	collectors := make([]*utils.DurationCollector, *flags.StoreWorkers)

	for i := uint(0); i < *flags.StoreWorkers; i++ {
		wg.Add(1)

		go func(index uint) {
			defer wg.Done()

			collector := utils.NewDurationCollector(int(*flags.StoreItemCount))

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

func RunStoreWorkersSplitRange(ctx context.Context, st store.Store, length uint, fn func(context.Context, store.Store, *utils.DurationCollector, uint, uint) error) *reporter.BenchmarkRun {
	workDivision := length / *flags.StoreWorkers

	return RunStoreWorkers(ctx, st, func(ctx context.Context, s store.Store, collector *utils.DurationCollector, u uint) error {
		end := workDivision * (u + 1)
		if end > length {
			end = length
		}

		return fn(ctx, st, collector, u*workDivision, end)
	})
}
