package utils

import (
	"fmt"
	"time"
)

// ScopedTimer tracks the duration between invocations to Start and Stop.
type ScopedTimer struct {
	start time.Time
	end   time.Time
}

func (s *ScopedTimer) Start() {
	s.start = time.Now()
}

func (s *ScopedTimer) Stop() {
	s.end = time.Now()
}

func (s *ScopedTimer) Elapsed() time.Duration {
	return s.end.Sub(s.start)
}

// TimingCalculator calculates the fastest , slowest and total duration of a given time period. Feed it duration
// information via the Sample function.
type TimingCalculator struct {
	Fastest     time.Duration
	Slowest     time.Duration
	Total       time.Duration
	SampleCount int64
}

func (t *TimingCalculator) Reset() {
	t.Fastest = 0 * time.Second
	t.Slowest = 0 * time.Second
	t.Total = 0 * time.Second
	t.SampleCount = 0
}

func (t *TimingCalculator) Sample(duration time.Duration) {
	if t.Fastest == 0 || duration < t.Fastest {
		t.Fastest = duration
	}

	if duration > t.Slowest {
		t.Slowest = duration
	}

	t.Total += duration
	t.SampleCount += 1
}

func (t *TimingCalculator) Merge(other *TimingCalculator) {
	if other.Fastest < t.Fastest {
		t.Fastest = other.Fastest
	}

	if other.Slowest > t.Slowest {
		t.Slowest = other.Slowest
	}

	t.Total += other.Total
	t.SampleCount += other.SampleCount
}

func (t *TimingCalculator) Average() time.Duration {
	if t.Total == 0 {
		return 0
	}

	return t.Total / time.Duration(t.SampleCount)
}

func (t *TimingCalculator) String() string {
	return fmt.Sprintf("Fastest:%06dms Slowest:%06dms Average:%06dms Total:%08dms SampleCount:%04d",
		t.Fastest.Milliseconds(), t.Slowest.Milliseconds(), t.Average().Milliseconds(), t.Total.Milliseconds(), t.SampleCount)
}
