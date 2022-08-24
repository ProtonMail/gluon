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
	storeCountFlag  = flag.Uint("imap-store-count", 0, "Total number of messages to store during store benchmarks.")
	storeListFlag   = flag.String("imap-store-list", "", "Use a list of predefined sequences to store rather than random generated.")
	storeSilentFlag = flag.Bool("imap-store-silent", false, "When set to true, request silent updates that do not produce any responses")
	storeAllFlag    = flag.Bool("imap-store-all", false, "If set, perform one store for all messages.")
)

type StoreBench struct {
	*stateTracker
	seqSets *ParallelSeqSet
}

func NewStore() benchmark.Benchmark {
	return NewIMAPBenchmarkRunner(&StoreBench{stateTracker: newStateTracker()})
}

func (*StoreBench) Name() string {
	return "imap-store"
}

func (s *StoreBench) Setup(ctx context.Context, addr net.Addr) error {
	return WithClient(addr, func(cl *client.Client) error {
		if _, err := s.createAndFillRandomMBox(cl); err != nil {
			return err
		}

		storeCount := uint32(*storeCountFlag)
		if storeCount == 0 {
			storeCount = uint32(*flags.IMAPMessageCount) / 2
		}

		seqSets, err := NewParallelSeqSet(storeCount,
			*flags.IMAPParallelClients,
			*storeListFlag,
			*storeAllFlag,
			*flags.IMAPRandomSeqSetIntervals,
			false,
			*flags.IMAPUIDMode)
		if err != nil {
			return err
		}

		s.seqSets = seqSets

		return nil
	})
}

func (s *StoreBench) TearDown(ctx context.Context, addr net.Addr) error {
	return s.cleanupWithAddr(addr)
}

func (s *StoreBench) Run(ctx context.Context, addr net.Addr) error {
	items := []string{"FLAGS", "-FLAGS", "+FLAGS"}
	flagList := []string{imap.DeletedFlag, imap.SeenFlag, imap.AnsweredFlag, imap.FlaggedFlag}

	RunParallelClientsWithMailbox(addr, s.MBoxes[0], false, func(cl *client.Client, index uint) {
		var storeFn func(*client.Client, *imap.SeqSet, int) error
		if *flags.IMAPUIDMode {
			storeFn = func(cl *client.Client, set *imap.SeqSet, index int) error {
				return UIDStore(cl, set, items[index%len(items)], *storeSilentFlag, flagList[index%len(flagList)])
			}
		} else {
			storeFn = func(cl *client.Client, set *imap.SeqSet, index int) error {
				return Store(cl, set, items[index%len(items)], *storeSilentFlag, flagList[index%len(flagList)])
			}
		}

		for s, v := range s.seqSets.Get(index) {
			if err := storeFn(cl, v, s); err != nil {
				panic(err)
			}
		}
	})

	return nil
}

func init() {
	benchmark.RegisterBenchmark(NewStore())
}
