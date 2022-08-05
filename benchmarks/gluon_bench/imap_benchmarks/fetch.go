package imap_benchmarks

import (
	"context"
	"flag"
	"net"

	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/benchmark"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/flags"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

var (
	fetchCountFlag = flag.Uint("fetch-count", 0, "Total number of messages to fetch during fetch benchmarks.")
	fetchListFlag  = flag.String("fetch-list", "", "Use a list of predefined sequences to fetch rather than random generated.")
	fetchReadOnly  = flag.Bool("fetch-read-only", false, "If set, perform fetches in read-only mode.")
	fetchAllFlag   = flag.Bool("fetch-all", false, "If set, perform one fetch for all messages.")
)

type Fetch struct {
	*stateTracker
	seqSets *ParallelSeqSet
}

func NewFetch() benchmark.Benchmark {
	return NewIMAPBenchmarkRunner(&Fetch{stateTracker: newStateTracker()})
}

func (*Fetch) Name() string {
	return "fetch"
}

func (f *Fetch) Setup(ctx context.Context, addr net.Addr) error {
	return WithClient(addr, func(cl *client.Client) error {
		if _, err := f.createAndFillRandomMBox(cl); err != nil {
			return err
		}

		fetchCount := uint32(*fetchCountFlag)
		if fetchCount == 0 {
			fetchCount = uint32(*flags.MessageCount) / 2
		}

		seqSets, err := NewParallelSeqSet(fetchCount,
			*flags.ParallelClients,
			*fetchListFlag,
			*fetchAllFlag,
			*flags.RandomSeqSetIntervals,
			false,
			*flags.UIDMode)

		if err != nil {
			return err
		}

		f.seqSets = seqSets
		return nil
	})
}

func (f *Fetch) TearDown(ctx context.Context, addr net.Addr) error {
	return f.cleanupWithAddr(addr)
}

func (f *Fetch) Run(ctx context.Context, addr net.Addr) error {
	RunParallelClientsWithMailbox(addr, f.MBoxes[0], *fetchReadOnly, func(cl *client.Client, index uint) {
		var fetchFn func(*client.Client, *imap.SeqSet) error
		if *flags.UIDMode {
			fetchFn = func(cl *client.Client, set *imap.SeqSet) error {
				return UIDFetchMessage(cl, set, imap.FetchAll)
			}
		} else {
			fetchFn = func(cl *client.Client, set *imap.SeqSet) error {
				return FetchMessage(cl, set, imap.FetchAll)
			}
		}

		for _, v := range f.seqSets.Get(index) {
			if err := fetchFn(cl, v); err != nil {
				panic(err)
			}
		}
	})

	return nil
}

func init() {
	benchmark.RegisterBenchmark(NewFetch())
}
