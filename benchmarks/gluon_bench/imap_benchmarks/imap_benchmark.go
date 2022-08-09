package imap_benchmarks

import (
	"context"
	"fmt"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/reporter"
	"github.com/ProtonMail/gluon/profiling"
	"net"
	"strings"
	"time"
)

// IMAPBenchmark is intended to be used to build benchmarks which bench IMAP commands on a given server.
type IMAPBenchmark interface {
	// Name should return the name of the benchmark. It will also be used to match against cli args.
	Name() string

	// Setup sets up the benchmark state, this is not timed.
	Setup(ctx context.Context, addr net.Addr) error

	// Run performs the actual benchmark, this is timed.
	Run(ctx context.Context, addr net.Addr) error

	// TearDown clear the benchmark state, this is not timed.
	TearDown(ctx context.Context, addr net.Addr) error
}

type IMAPBenchmarkExtra struct {
	CMDStatistic [profiling.CmdTypeTotal]*reporter.BenchmarkStatistics
}

func (i *IMAPBenchmarkExtra) String() string {
	builder := strings.Builder{}

	for n, v := range i.CMDStatistic {
		if v.SampleCount == 0 {
			continue
		}

		builder.WriteString(fmt.Sprintf("[%v] %v\n", profiling.CmdTypeToString(n), v.String()))
	}

	return builder.String()
}

func NewIMAPBenchmarkRun(duration time.Duration, cmdTimings [profiling.CmdTypeTotal][]time.Duration) *reporter.BenchmarkRun {
	var cmdStatistic [profiling.CmdTypeTotal]*reporter.BenchmarkStatistics
	for i, v := range cmdTimings {
		cmdStatistic[i] = reporter.NewBenchmarkStatistics(nil, v...)
	}

	return reporter.NewBenchmarkRunSingle(duration, &IMAPBenchmarkExtra{CMDStatistic: cmdStatistic})
}
