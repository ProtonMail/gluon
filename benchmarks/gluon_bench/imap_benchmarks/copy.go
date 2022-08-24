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
	copyCountFlag = flag.Uint("imap-copy-count", 0, "Total number of messages to copy during copy benchmarks.")
	copyListFlag  = flag.String("imap-copy-list", "", "Use a list of predefined sequences to copy rather than random generated.")
	copyAllFlag   = flag.Bool("imap-copy-all", false, "If set, perform a copy of the all messages.")
)

type Copy struct {
	*stateTracker
	seqSets *ParallelSeqSet
}

func NewCopy() benchmark.Benchmark {
	return NewIMAPBenchmarkRunner(&Copy{
		stateTracker: newStateTracker(),
	})
}

func (*Copy) Name() string {
	return "imap-copy"
}

func (c *Copy) Setup(ctx context.Context, addr net.Addr) error {
	return WithClient(addr, func(cl *client.Client) error {
		if _, err := c.createAndFillRandomMBox(cl); err != nil {
			return err
		}

		if _, err := c.createRandomMBox(cl); err != nil {
			return err
		}

		copyCount := uint32(*copyCountFlag)
		if copyCount == 0 {
			copyCount = uint32(*flags.IMAPMessageCount / 2)
		}

		seqSets, err := NewParallelSeqSet(copyCount,
			*flags.IMAPParallelClients,
			*copyListFlag,
			*copyAllFlag,
			*flags.IMAPRandomSeqSetIntervals,
			false,
			*flags.IMAPUIDMode)
		if err != nil {
			return err
		}

		c.seqSets = seqSets

		return nil
	})
}

func (c *Copy) TearDown(ctx context.Context, addr net.Addr) error {
	return c.cleanupWithAddr(addr)
}

func (c *Copy) Run(ctx context.Context, addr net.Addr) error {
	srcMBox := c.MBoxes[0]
	dstMBox := c.MBoxes[1]

	RunParallelClientsWithMailbox(addr, srcMBox, true, func(cl *client.Client, index uint) {
		var copyFn func(*client.Client, *imap.SeqSet, string) error
		if *flags.IMAPUIDMode {
			copyFn = func(cl *client.Client, set *imap.SeqSet, mailbox string) error {
				return cl.UidCopy(set, mailbox)
			}
		} else {
			copyFn = func(cl *client.Client, set *imap.SeqSet, mailbox string) error {
				return cl.Copy(set, mailbox)
			}
		}

		for _, v := range c.seqSets.Get(index) {
			if err := copyFn(cl, v, dstMBox); err != nil {
				panic(err)
			}
		}
	})

	return nil
}

func init() {
	benchmark.RegisterBenchmark(NewCopy())
}
