package reporter

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/bradenaw/juniper/xslices"
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
	Extra        BenchmarkExtra
}

func (b *BenchmarkStatistics) String() string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("SampleCount:%04d Total:%v Fastest:%v Slowest:%v Average:%v Median:%v 90thPercentile:%v 10thPercentile:%v RMS:%v",
		b.SampleCount, b.Total, b.Fastest, b.Slowest, b.Average,
		b.Median, b.Percentile90, b.Percentile10, b.RMS,
	))

	if b.Extra != nil {
		builder.WriteString(" Extra:\n")
		builder.WriteString(b.Extra.String())
	}

	return builder.String()
}

func NewBenchmarkStatistics(extra BenchmarkExtra, durations ...time.Duration) *BenchmarkStatistics {
	sortedDurations := durations
	sort.Slice(sortedDurations, func(i1, i2 int) bool {
		return sortedDurations[i1] < sortedDurations[i2]
	})

	statistics := &BenchmarkStatistics{
		Extra: extra,
	}
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
			statistics.Median = (sortedDurations[halfPoint-1] + sortedDurations[halfPoint]) / 2
		} else {
			statistics.Median = sortedDurations[((statistics.SampleCount+1)/2)-1]
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

type BenchmarkExtra interface {
	String() string
}

type BenchmarkRun struct {
	Durations []time.Duration
	Extra     BenchmarkExtra
}

func NewBenchmarkRunSingle(duration time.Duration, extra BenchmarkExtra) *BenchmarkRun {
	return &BenchmarkRun{Durations: []time.Duration{duration}, Extra: extra}
}

func NewBenchmarkRun(durations []time.Duration, extra BenchmarkExtra) *BenchmarkRun {
	return &BenchmarkRun{Durations: durations, Extra: extra}
}

type BenchmarkReport struct {
	Name       string
	Runs       []*BenchmarkStatistics
	Statistics *BenchmarkStatistics
}

func NewBenchmarkReport(name string, runs ...*BenchmarkStatistics) *BenchmarkReport {
	durations := xslices.Map(runs, func(r *BenchmarkStatistics) time.Duration {
		return r.Total
	})

	return &BenchmarkReport{Name: name, Runs: runs, Statistics: NewBenchmarkStatistics(nil, durations...)}
}

// BenchmarkReporter is the interface that is required to be implemented by any report generation tool.
type BenchmarkReporter interface {
	ProduceReport(reports []*BenchmarkReport) error
}
