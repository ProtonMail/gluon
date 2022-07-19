package utils

import (
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
