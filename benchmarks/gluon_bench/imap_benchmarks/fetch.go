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
	fetchCountFlag = flag.Uint("imap-fetch-count", 0, "Total number of messages to fetch during fetch benchmarks.")
	fetchListFlag  = flag.String("imap-fetch-list", "", "Use a list of predefined sequences to fetch rather than random generated.")
	fetchReadOnly  = flag.Bool("imap-fetch-read-only", false, "If set, perform fetches in read-only mode.")
	fetchAllFlag   = flag.Bool("imap-fetch-all", false, "If set, perform one fetch for all messages.")
)

type Fetch struct {
	*stateTracker
	seqSets *ParallelSeqSet
}

func NewFetch() benchmark.Benchmark {
	return NewIMAPBenchmarkRunner(&Fetch{stateTracker: newStateTracker()})
}

func (*Fetch) Name() string {
	return "imap-fetch"
}

func (f *Fetch) Setup(ctx context.Context, addr net.Addr) error {
	return WithClient(addr, func(cl *client.Client) error {
		if _, err := f.createAndFillRandomMBox(cl); err != nil {
			return err
		}

		fetchCount := uint32(*fetchCountFlag)
		if fetchCount == 0 {
			fetchCount = uint32(*flags.IMAPMessageCount) / 2
		}

		seqSets, err := NewParallelSeqSet(fetchCount,
			*flags.IMAPParallelClients,
			*fetchListFlag,
			*fetchAllFlag,
			*flags.IMAPRandomSeqSetIntervals,
			false,
			*flags.IMAPUIDMode)
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
	attributes := []imap.FetchItem{
		imap.FetchFlags,
		imap.FetchRFC822Size,
		imap.FetchRFC822Header,
		imap.FetchInternalDate,
		imap.FetchRFC822,
		imap.FetchRFC822Text,
		imap.FetchBody,
		"BODY[]",
	}

	RunParallelClientsWithMailbox(addr, f.MBoxes[0], *fetchReadOnly, func(cl *client.Client, index uint) {
		var fetchFn func(*client.Client, *imap.SeqSet) error
		if *flags.IMAPUIDMode {
			fetchFn = func(cl *client.Client, set *imap.SeqSet) error {
				return UIDFetchMessage(cl, set, attributes...)
			}
		} else {
			fetchFn = func(cl *client.Client, set *imap.SeqSet) error {
				return FetchMessage(cl, set, attributes...)
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
