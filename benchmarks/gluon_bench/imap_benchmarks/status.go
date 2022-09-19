package imap_benchmarks

import (
	"context"
	"flag"
	"net"

	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/benchmark"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

var statusCallCountFlag = flag.Uint("imap-status-count", 1000, "Number of times to call status.")

type Status struct {
	*stateTracker
}

func NewStatus() benchmark.Benchmark {
	return NewIMAPBenchmarkRunner(&Status{stateTracker: newStateTracker()})
}

func (*Status) Name() string {
	return "imap-status"
}

func (s *Status) Setup(ctx context.Context, addr net.Addr) error {
	return WithClient(addr, func(cl *client.Client) error {
		if _, err := s.createAndFillRandomMBox(cl); err != nil {
			return err
		}

		return nil
	})
}

func (s *Status) TearDown(ctx context.Context, addr net.Addr) error {
	return s.cleanupWithAddr(addr)
}

func (s *Status) Run(ctx context.Context, addr net.Addr) error {
	RunParallelClientsWithMailbox(addr, s.MBoxes[0], *fetchReadOnly, func(cl *client.Client, index uint) {
		for i := uint(0); i < *statusCallCountFlag; i++ {
			_, err := cl.Status(s.MBoxes[0], []imap.StatusItem{imap.StatusRecent, imap.StatusMessages, imap.StatusRecent, imap.StatusUnseen, imap.StatusUidNext, imap.StatusUidValidity})
			if err != nil {
				panic(err)
			}
		}
	})

	return nil
}

func init() {
	benchmark.RegisterBenchmark(NewStatus())
}
