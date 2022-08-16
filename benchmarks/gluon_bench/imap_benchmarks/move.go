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
	moveListFlag        = flag.String("imap-move-list", "", "Use a list of predefined sequences to move rather than random generated.")
	moveAllFlag         = flag.Bool("imap-move-all", false, "If set, perform a move of the all messages.")
	moveIntoSameDstFlag = flag.Bool("imap-move-into-same-dst", false, "If set, rather than moving each unique mailbox into separate unique mailboxes, move all messages into one common destination mailbox.")
)

type Move struct {
	*stateTracker
	seqSets       *ParallelSeqSet
	messageCounts []uint32
	srcMailboxes  []string
	dstMailboxes  []string
}

func NewMove() benchmark.Benchmark {
	return NewIMAPBenchmarkRunner(&Move{stateTracker: newStateTracker()})
}

func (*Move) Name() string {
	return "imap-move"
}

func (m *Move) Setup(ctx context.Context, addr net.Addr) error {
	if *flags.IMAPMessageCount == 0 {
		return fmt.Errorf("move benchmark requires a message count > 0")
	}

	return WithClient(addr, func(cl *client.Client) error {
		m.srcMailboxes = make([]string, 0, *flags.IMAPParallelClients)
		m.dstMailboxes = make([]string, 0, *flags.IMAPParallelClients)

		for i := uint(0); i < *flags.IMAPParallelClients; i++ {
			mbox, err := m.createAndFillRandomMBox(cl)
			if err != nil {
				return err
			}

			m.srcMailboxes = append(m.srcMailboxes, mbox)
		}

		var dstMboxCount uint
		if *moveIntoSameDstFlag {
			dstMboxCount = 1
		} else {
			dstMboxCount = *flags.IMAPParallelClients
		}

		for i := uint(0); i < dstMboxCount; i++ {
			mbox, err := m.createRandomMBox(cl)
			if err != nil {
				return err
			}

			m.dstMailboxes = append(m.dstMailboxes, mbox)
		}

		seqSets, err := NewParallelSeqSet(uint32(*flags.IMAPMessageCount),
			*flags.IMAPParallelClients,
			*moveListFlag,
			*moveAllFlag,
			*flags.IMAPRandomSeqSetIntervals,
			true,
			*flags.IMAPUIDMode)
		if err != nil {
			return err
		}

		m.seqSets = seqSets

		return nil
	})
}

func (m *Move) TearDown(ctx context.Context, addr net.Addr) error {
	return m.cleanupWithAddr(addr)
}

func (m *Move) Run(ctx context.Context, addr net.Addr) error {
	mboxInfos := xslices.Map(m.srcMailboxes, func(name string) MailboxInfo {
		return MailboxInfo{
			Name:     name,
			ReadOnly: true,
		}
	})

	RunParallelClientsWithMailboxes(addr, mboxInfos, func(cl *client.Client, index uint) {
		var moveFn func(*client.Client, *imap.SeqSet, string) error
		if *flags.IMAPUIDMode {
			moveFn = func(cl *client.Client, set *imap.SeqSet, mailbox string) error {
				return cl.UidMove(set, mailbox)
			}
		} else {
			moveFn = func(cl *client.Client, set *imap.SeqSet, mailbox string) error {
				return cl.Move(set, mailbox)
			}
		}

		for _, v := range m.seqSets.Get(index) {
			if *moveIntoSameDstFlag {
				if err := moveFn(cl, v, m.dstMailboxes[0]); err != nil {
					panic(err)
				}
			} else {
				if err := moveFn(cl, v, m.dstMailboxes[index]); err != nil {
					panic(err)
				}
			}
		}
	})

	return nil
}

func init() {
	benchmark.RegisterBenchmark(NewMove())
}
