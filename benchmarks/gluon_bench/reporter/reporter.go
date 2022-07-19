package reporter

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/bradenaw/juniper/xslices"

	"github.com/ProtonMail/gluon/profiling"
)

type BenchmarkStatistics struct {
	Total        time.Duration
	Average      time.Duration
	Fastest      time.Duration
	Slowest      time.Duration
	Median       time.Duration
	Percentile90 time.Duration
	Percentile10 time.Duration
	RMS          time.Duration
	SampleCount  int
}

func (b *BenchmarkStatistics) String() string {
	return fmt.Sprintf("Total: %04dms SampleCount:%04d Fastest:%04dms Slowest:%04dms Average: %04dms Median: %04dms 90thPercentile: %04d ms 10thPercentile: %04dms RMS %04dms",
		b.Total.Milliseconds(), b.SampleCount, b.Fastest.Milliseconds(), b.Slowest.Milliseconds(), b.Average.Milliseconds(),
		b.Median.Milliseconds(), b.Percentile90.Milliseconds(), b.Percentile10.Milliseconds(), b.RMS.Milliseconds(),
	)
}

func NewBenchmarkStatistics(durations ...time.Duration) BenchmarkStatistics {
	sortedDurations := durations
	sort.Slice(sortedDurations, func(i1, i2 int) bool {
		return sortedDurations[i1] < sortedDurations[i2]
	})

	statistics := BenchmarkStatistics{}
	statistics.SampleCount = len(sortedDurations)

	if statistics.SampleCount == 1 {
		statistics.Fastest = sortedDurations[0]
		statistics.Slowest = sortedDurations[0]
		statistics.Average = sortedDurations[0]
		statistics.Total = sortedDurations[0]
		statistics.Median = sortedDurations[0]
		statistics.Percentile90 = sortedDurations[0]
		statistics.Percentile10 = sortedDurations[0]
		statistics.RMS = sortedDurations[0]
	} else if statistics.SampleCount > 1 {
		statistics.Fastest = sortedDurations[0]
		statistics.Slowest = sortedDurations[statistics.SampleCount-1]
		statistics.Total = xslices.Reduce(sortedDurations, 0, func(v1 time.Duration, v2 time.Duration) time.Duration {
			return v1 + v2
		})
		statistics.Average = statistics.Total / time.Duration(statistics.SampleCount)
		if statistics.Total%2 == 0 {
			halfPoint := statistics.SampleCount / 2
			statistics.Median = (sortedDurations[halfPoint] + sortedDurations[halfPoint+1]) / 2
		} else {
			statistics.Median = sortedDurations[((statistics.SampleCount + 1) / 2)]
		}
		statistics.Percentile90 = sortedDurations[int(math.Floor(float64(statistics.SampleCount)*(90.0/100.0)))]
		statistics.Percentile10 = sortedDurations[int(math.Floor(float64(statistics.SampleCount)*(10.0/100.0)))]
		var sumSquaredWithDiv float64
		for i := 0; i < statistics.SampleCount; i++ {
			// Dividing now rather than later or else we will trigger overflow.
			f64Duration := float64(sortedDurations[i])
			sumSquaredWithDiv += (f64Duration * f64Duration) / float64(statistics.SampleCount)
		}
		statistics.RMS = time.Duration(math.Round(math.Sqrt(sumSquaredWithDiv)))
	}

	return statistics
}

type BenchmarkRun struct {
	Duration      time.Duration
	CmdStatistics [profiling.CmdTypeTotal]BenchmarkStatistics
}

func NewBenchmarkRun(duration time.Duration, cmdTimings [profiling.CmdTypeTotal][]time.Duration) *BenchmarkRun {
	var cmdStatistic [profiling.CmdTypeTotal]BenchmarkStatistics
	for i, v := range cmdTimings {
		cmdStatistic[i] = NewBenchmarkStatistics(v...)
	}

	return &BenchmarkRun{Duration: duration, CmdStatistics: cmdStatistic}
}

type BenchmarkReport struct {
	Name       string
	Runs       []*BenchmarkRun
	Statistics BenchmarkStatistics
}

func NewBenchmarkReport(name string, runs ...*BenchmarkRun) *BenchmarkReport {
	durations := xslices.Map(runs, func(r *BenchmarkRun) time.Duration {
		return r.Duration
	})

	return &BenchmarkReport{Name: name, Runs: runs, Statistics: NewBenchmarkStatistics(durations...)}
}

// BenchmarkReporter is the interface that is required to be implemented by any report generation tool.
type BenchmarkReporter interface {
	ProduceReport(reports []*BenchmarkReport) error
}
