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
	"github.com/google/uuid"
)

var (
	expungeCountFlag    = flag.Uint("expunge-count", 0, "Total number of messages to expunge during expunge benchmarks.")
	expungeSameMBoxFlag = flag.Bool("expunge-same-mbox", false, "When true run all the expunge test on the same inbox rather than separate ones in parallel.")
	expungeListFlag     = flag.String("expunge-list", "", "Use a list of predefined sequences to expunge rather than random generated. Only works when -expunge-same-mbox is not set.")
	expungeAllFlag      = flag.Bool("expunge-all", false, "If set, perform a expunge of the all messages. Only works when -expunge-same-mbox is not set.")
)

type Expunge struct {
	seqSets   *ParallelSeqSet
	mailboxes []string
}

func NewExpunge() benchmark.Benchmark {
	return NewIMAPBenchmarkRunner(&Expunge{})
}

func (*Expunge) Name() string {
	return "expunge"
}

func (e *Expunge) Setup(ctx context.Context, addr net.Addr) error {
	cl, err := NewClient(addr.String())
	if err != nil {
		return err
	}

	defer CloseClient(cl)

	if *expungeSameMBoxFlag {
		if err := FillBenchmarkSourceMailbox(cl); err != nil {
			return err
		}

		status, err := cl.Status(*flags.Mailbox, []imap.StatusItem{imap.StatusMessages})
		if err != nil {
			return err
		}

		messageCount := status.Messages

		if messageCount == 0 {
			return fmt.Errorf("mailbox '%v' has no messages", *flags.Mailbox)
		}

		expungeCount := uint32(*expungeCountFlag)
		if expungeCount == 0 {
			expungeCount = messageCount / 2
		}

		e.seqSets = NewParallelSeqSetExpunge(expungeCount,
			*flags.ParallelClients,
			*flags.RandomSeqSetIntervals,
			*flags.UIDMode,
		)

		for i := uint(0); i < *flags.ParallelClients; i++ {
			e.mailboxes = append(e.mailboxes, *flags.Mailbox)
		}
	} else {
		e.mailboxes = make([]string, 0, *flags.ParallelClients)
		for i := uint(0); i < *flags.ParallelClients; i++ {
			e.mailboxes = append(e.mailboxes, uuid.NewString())
		}

		for _, v := range e.mailboxes {
			if err := cl.Create(v); err != nil {
				return err
			}

			if err := BuildMailbox(cl, v, int(*flags.FillSourceMailbox)); err != nil {
				return err
			}
		}

		seqSets, err := NewParallelSeqSet(uint32(*flags.FillSourceMailbox),
			*flags.ParallelClients,
			*expungeListFlag,
			*expungeAllFlag,
			*flags.RandomSeqSetIntervals,
			true,
			*flags.UIDMode)

		if err != nil {
			return err
		}

		e.seqSets = seqSets
	}

	return nil
}

func (e *Expunge) TearDown(ctx context.Context, addr net.Addr) error {
	cl, err := NewClient(addr.String())
	if err != nil {
		return err
	}

	defer CloseClient(cl)

	if !*expungeSameMBoxFlag {
		for _, v := range e.mailboxes {
			if err := cl.Delete(v); err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *Expunge) Run(ctx context.Context, addr net.Addr) error {
	mboxInfo := xslices.Map(e.mailboxes, func(m string) MailboxInfo {
		return MailboxInfo{Name: m, ReadOnly: false}
	})

	RunParallelClientsWithMailboxes(addr, mboxInfo, func(cl *client.Client, index uint) {
		var expungeFn func(*client.Client, *imap.SeqSet) error
		if *flags.UIDMode {
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
