package timing

import "time"

// Timer tracks the duration between invocations to Start and Stop.
type Timer struct {
	start time.Time
	end   time.Time
}

func (s *Timer) Start() {
	s.start = time.Now()
}

func (s *Timer) Stop() {
	s.end = time.Now()
}

func (s *Timer) Elapsed() time.Duration {
	return s.end.Sub(s.start)
}

type Collector struct {
	durations []time.Duration
	timer     Timer
}

func NewDurationCollector(capacity int) *Collector {
	return &Collector{
		durations: make([]time.Duration, 0, capacity),
	}
}

func (d *Collector) Start() {
	d.timer.Start()
}

func (d *Collector) Stop() {
	d.timer.Stop()
	d.durations = append(d.durations, d.timer.Elapsed())
}

func (d *Collector) Durations() []time.Duration {
	return d.durations
}
