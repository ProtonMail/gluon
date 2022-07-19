package reporter

import (
	"time"

	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/utils"
	"github.com/ProtonMail/gluon/profiling"
)

type BenchmarkRun struct {
	Duration   time.Duration
	CmdTimings [profiling.CmdTypeTotal]utils.TimingCalculator
}

func NewBenchmarkRun(duration time.Duration, cmdTimings [profiling.CmdTypeTotal]utils.TimingCalculator) *BenchmarkRun {
	return &BenchmarkRun{Duration: duration, CmdTimings: cmdTimings}
}

type BenchmarkReport struct {
	Name string
	Runs []*BenchmarkRun
}

func NewBenchmarkReport(name string, runs ...*BenchmarkRun) *BenchmarkReport {
	return &BenchmarkReport{Name: name, Runs: runs}
}

// BenchmarkReporter is the interface that is required to be implemented by any report generation tool.
type BenchmarkReporter interface {
	ProduceReport(reports []*BenchmarkReport) error
}
