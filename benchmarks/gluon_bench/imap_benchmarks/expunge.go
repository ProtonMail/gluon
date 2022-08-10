package imap_benchmarks

import (
	"context"
	"flag"
	"fmt"
	"net"

	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/benchmark"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/flags"
	"github.com/bradenaw/juniper/xslices"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

var (
	expungeCountFlag    = flag.Uint("imap-expunge-count", 0, "Total number of messages to expunge during expunge benchmarks.")
	expungeSameMBoxFlag = flag.Bool("imap-expunge-same-mbox", false, "When true run all the expunge test on the same inbox rather than separate ones in parallel.")
	expungeListFlag     = flag.String("imap-expunge-list", "", "Use a list of predefined sequences to expunge rather than random generated. Only works when -expunge-same-mbox is not set.")
	expungeAllFlag      = flag.Bool("imap-expunge-all", false, "If set, perform a expunge of the all messages. Only works when -expunge-same-mbox is not set.")
)

type Expunge struct {
	*stateTracker
	seqSets  *ParallelSeqSet
	mboxInfo []MailboxInfo
}

func NewExpunge() benchmark.Benchmark {
	return NewIMAPBenchmarkRunner(&Expunge{stateTracker: newStateTracker()})
}

func (*Expunge) Name() string {
	return "imap-expunge"
}

func (e *Expunge) Setup(ctx context.Context, addr net.Addr) error {
	return WithClient(addr, func(cl *client.Client) error {
		if *expungeSameMBoxFlag {
			if _, err := e.createAndFillRandomMBox(cl); err != nil {
				return nil
			}

			expungeCount := uint32(*expungeCountFlag)
			if expungeCount == 0 {
				expungeCount = uint32(*flags.IMAPMessageCount) / 2
			}

			e.seqSets = NewParallelSeqSetExpunge(expungeCount,
				*flags.IMAPParallelClients,
				*flags.IMAPRandomSeqSetIntervals,
				*flags.IMAPUIDMode,
			)

			e.mboxInfo = make([]MailboxInfo, *flags.IMAPParallelClients)
			for i := 0; i < len(e.mboxInfo); i++ {
				e.mboxInfo[i] = MailboxInfo{Name: e.MBoxes[0], ReadOnly: false}
			}
		} else {
			for i := uint(0); i < *flags.IMAPParallelClients; i++ {
				if _, err := e.createAndFillRandomMBox(cl); err != nil {
					return err
				}
			}

			seqSets, err := NewParallelSeqSet(uint32(*flags.IMAPMessageCount),
				*flags.IMAPParallelClients,
				*expungeListFlag,
				*expungeAllFlag,
				*flags.IMAPRandomSeqSetIntervals,
				true,
				*flags.IMAPUIDMode)

			if err != nil {
				return err
			}

			e.seqSets = seqSets

			e.mboxInfo = xslices.Map(e.MBoxes, func(m string) MailboxInfo {
				return MailboxInfo{Name: m, ReadOnly: false}
			})
		}
		return nil
	})
}

func (e *Expunge) TearDown(ctx context.Context, addr net.Addr) error {
	return e.cleanupWithAddr(addr)
}

func (e *Expunge) Run(ctx context.Context, addr net.Addr) error {
	RunParallelClientsWithMailboxes(addr, e.mboxInfo, func(cl *client.Client, index uint) {
		var expungeFn func(*client.Client, *imap.SeqSet) error
		if *flags.IMAPUIDMode {
			expungeFn = func(cl *client.Client, set *imap.SeqSet) error {
				if err := UIDStore(cl, set, "+FLAGS", true, imap.DeletedFlag); err != nil {
					return err
				}
				return cl.Expunge(nil)
			}
		} else {
			expungeFn = func(cl *client.Client, set *imap.SeqSet) error {
				if err := Store(cl, set, "+FLAGS", true, imap.DeletedFlag); err != nil {
					return err
				}
				return cl.Expunge(nil)
			}
		}

		for _, v := range e.seqSets.Get(index) {
			if err := expungeFn(cl, v); err != nil {
				panic(fmt.Sprintf("Seq:%v err:%v", v, err))
			}
		}
	})

	return nil
}

func init() {
	benchmark.RegisterBenchmark(NewExpunge())
}
