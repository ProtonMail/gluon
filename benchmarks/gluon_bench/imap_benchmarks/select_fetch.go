package imap_benchmarks

import (
	"context"
	"flag"
	"github.com/emersion/go-imap"
	"net"

	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/benchmark"
	"github.com/bradenaw/juniper/xslices"
	"github.com/emersion/go-imap/client"
)

var (
	selectFetchRepetitionsFlag = flag.Uint("imap-select-fetch-repeat", 50, "Number of times to repeat the request.")
)

type SelectFetch struct {
	*stateTracker
	mboxInfo []MailboxInfo
}

func NewSelectFetch() benchmark.Benchmark {
	return NewIMAPBenchmarkRunner(&SelectFetch{stateTracker: newStateTracker()})
}

func (*SelectFetch) Name() string {
	return "imap-select-fetch"
}

func (e *SelectFetch) Setup(ctx context.Context, addr net.Addr) error {
	return WithClient(addr, func(cl *client.Client) error {
		for i := uint(0); i < *selectFetchRepetitionsFlag; i++ {
			if _, err := e.createAndFillRandomMBox(cl); err != nil {
				return err
			}
		}

		e.mboxInfo = xslices.Map(e.MBoxes, func(m string) MailboxInfo {
			return MailboxInfo{Name: m, ReadOnly: false}
		})
		return nil
	})
}

func (e *SelectFetch) TearDown(ctx context.Context, addr net.Addr) error {
	return e.cleanupWithAddr(addr)
}

func (e *SelectFetch) Run(ctx context.Context, addr net.Addr) error {
	RunParallelClients(addr, func(cl *client.Client, u uint) {
		for i := uint(0); i < *selectFetchRepetitionsFlag; i++ {
			if _, err := cl.Select(e.mboxInfo[i].Name, e.mboxInfo[i].ReadOnly); err != nil {
				panic(err)
			}

			if err := FetchMessage(cl, NewSequenceSetAll(), imap.FetchUid, imap.FetchFlags); err != nil {
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
	benchmark.RegisterBenchmark(NewSelectFetch())
}
