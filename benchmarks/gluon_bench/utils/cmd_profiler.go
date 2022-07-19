package utils

import (
	"sync"
	"time"

	"github.com/ProtonMail/gluon/profiling"
)

// DurationCmdProfiler records the duration of the duration between invocations of IMAP Commands.
type DurationCmdProfiler struct {
	calculator [profiling.CmdTypeTotal]TimingCalculator
	start      [profiling.CmdTypeTotal]time.Time
}

func (c *DurationCmdProfiler) Start(cmdType int) {
	// We can use time since Go 1.9 they have switch to monotonic clocks.
	c.start[cmdType] = time.Now()
}

func (c *DurationCmdProfiler) Stop(cmdType int) {
	elapsed := time.Now().Sub(c.start[cmdType])
	c.calculator[cmdType].Sample(elapsed)
}

func NewDurationCmdProfiler() *DurationCmdProfiler {
	return &DurationCmdProfiler{}
}

type DurationCmdProfilerBuilder struct {
	mutex     sync.Mutex
	profilers []*DurationCmdProfiler
}

func (c *DurationCmdProfilerBuilder) New() profiling.CmdProfiler {
	profiler := NewDurationCmdProfiler()
	c.profilers = append(c.profilers, profiler)

	return profiler
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
func (c *DurationCmdProfilerBuilder) Merge() [profiling.CmdTypeTotal]TimingCalculator {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var result [profiling.CmdTypeTotal]TimingCalculator

	for _, v := range c.profilers {
		for i := 0; i < len(result); i++ {
			result[i].Merge(&v.calculator[i])
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
