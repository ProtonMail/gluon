package imap_benchmarks

import (
	"context"
	"fmt"
	"net"

	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/benchmark"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/flags"
	"github.com/emersion/go-imap/client"
)

type Append struct {
	*stateTracker
}

func NewAppend() benchmark.Benchmark {
	return NewIMAPBenchmarkRunner(&Append{stateTracker: newStateTracker()})
}

func (a *Append) Name() string {
	return "imap-append"
}

func (a *Append) Setup(ctx context.Context, addr net.Addr) error {
	if *flags.IMAPMessageCount == 0 {
		return fmt.Errorf("invalid message count")
	}

	return WithClient(addr, func(cl *client.Client) error {
		for i := uint(0); i < *flags.IMAPParallelClients; i++ {
			if _, err := a.createRandomMBox(cl); err != nil {
				return err
			}
		}

		return nil
	})
}

func (a *Append) TearDown(ctx context.Context, addr net.Addr) error {
	//return a.cleanupWithAddr(addr)
	return nil
}

func (a *Append) Run(ctx context.Context, addr net.Addr) error {
	RunParallelClients(addr, func(c *client.Client, u uint) {
		if err := BuildMailbox(c, a.MBoxes[u], int(*flags.IMAPMessageCount)); err != nil {
			panic(err)
		}
	})

	return nil
}

func init() {
	benchmark.RegisterBenchmark(NewAppend())
}
