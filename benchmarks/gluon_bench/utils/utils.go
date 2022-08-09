package utils

import (
	"bufio"
	"os"
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

func ReadLinesFromFile(path string) ([]string, error) {
	readFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer readFile.Close()

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	lines := make([]string, 0, 16)

	for fileScanner.Scan() {
		lines = append(lines, fileScanner.Text())
	}

	return lines, nil
}

type DurationCollector struct {
	durations []time.Duration
	timer     ScopedTimer
}

func NewDurationCollector(capacity int) *DurationCollector {
	return &DurationCollector{
		durations: make([]time.Duration, 0, capacity),
	}
}

func (d *DurationCollector) Start() {
	d.timer.Start()
}

func (d *DurationCollector) Stop() {
	d.timer.Stop()
	d.durations = append(d.durations, d.timer.Elapsed())
}

func (d *DurationCollector) Durations() []time.Duration {
	return d.durations
}
