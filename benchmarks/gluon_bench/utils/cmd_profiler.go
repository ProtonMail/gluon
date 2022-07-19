package utils

import (
	"sync"
	"time"

	"github.com/ProtonMail/gluon/profiling"
)

// DurationCmdProfiler records the duration of the duration between invocations of IMAP Commands.
type DurationCmdProfiler struct {
	durations [profiling.CmdTypeTotal][]time.Duration
	start     [profiling.CmdTypeTotal]time.Time
}

func (c *DurationCmdProfiler) Start(cmdType int) {
	// We can use time since Go 1.9 they have switch to monotonic clocks.
	c.start[cmdType] = time.Now()
}

func (c *DurationCmdProfiler) Stop(cmdType int) {
	elapsed := time.Now().Sub(c.start[cmdType])
	c.durations[cmdType] = append(c.durations[cmdType], elapsed)
}

func NewDurationCmdProfiler() *DurationCmdProfiler {
	profiler := &DurationCmdProfiler{}
	for i := 0; i < len(profiler.durations); i++ {
		profiler.durations[i] = make([]time.Duration, 0, 128)
	}

	return profiler
}

type DurationCmdProfilerBuilder struct {
	mutex     sync.Mutex
	profilers []*DurationCmdProfiler
}

func (c *DurationCmdProfilerBuilder) New() profiling.CmdProfiler {
	return NewDurationCmdProfiler()
}

func (c *DurationCmdProfilerBuilder) Collect(profiler profiling.CmdProfiler) {
	switch v := profiler.(type) {
	case *DurationCmdProfiler:
		c.mutex.Lock()
		defer c.mutex.Unlock()

		c.profilers = append(c.profilers, v)
	}
}

// Merge merges all collected command profilers into a single timing Calculator for each IMAP command.
func (c *DurationCmdProfilerBuilder) Merge() [profiling.CmdTypeTotal][]time.Duration {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var result [profiling.CmdTypeTotal][]time.Duration

	for _, v := range c.profilers {
		for i := 0; i < len(result); i++ {
			result[i] = append(result[i], v.durations[i]...)
		}
	}

	return result
}

func (c *DurationCmdProfilerBuilder) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.profilers = nil
}

func NewDurationCmdProfilerBuilder() *DurationCmdProfilerBuilder {
	return &DurationCmdProfilerBuilder{}
}
