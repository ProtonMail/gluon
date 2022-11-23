package imap_benchmarks

import (
	"context"
	"flag"
	"net"

	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/benchmark"
	"github.com/emersion/go-imap/client"
)

var (
	selectReadOnlyFlag  = flag.Bool("imap-select-readonly", false, "If set to true, perform a read only select (examine).")
	selectCallCountFlag = flag.Uint("imap-select-count", 1000, "Number of times to call select.")
)

type Select struct {
	*stateTracker
}

func NewSelect() benchmark.Benchmark {
	return NewIMAPBenchmarkRunner(&Select{stateTracker: newStateTracker()})
}

func (*Select) Name() string {
	return "imap-select"
}

func (s *Select) Setup(ctx context.Context, addr net.Addr) error {
	return WithClient(addr, func(cl *client.Client) error {
		if _, err := s.createAndFillRandomMBox(cl); err != nil {
			return err
		}

		return nil
	})
}

func (s *Select) TearDown(ctx context.Context, addr net.Addr) error {
	return s.cleanupWithAddr(addr)
}

func (s *Select) Run(ctx context.Context, addr net.Addr) error {
	RunParallelClientsWithMailbox(addr, s.MBoxes[0], *fetchReadOnly, func(cl *client.Client, index uint) {
		for i := uint(0); i < *selectCallCountFlag; i++ {
			_, err := cl.Select(s.MBoxes[0], *selectReadOnlyFlag)
			if err != nil {
				panic(err)
			}

			if err := cl.Unselect(); err != nil {
				panic(err)
			}
		}
	})

	return nil
}

func init() {
	benchmark.RegisterBenchmark(NewSelect())
}
