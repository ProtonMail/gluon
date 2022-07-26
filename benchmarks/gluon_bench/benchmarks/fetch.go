package benchmarks

import (
	"context"
	"flag"
	"fmt"
	"net"

	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/flags"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/utils"
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
	seqSets *ParallelSeqSet
}

func NewFetch() *Fetch {
	return &Fetch{}
}

func (*Fetch) Name() string {
	return "fetch"
}

func (f *Fetch) Setup(ctx context.Context, addr net.Addr) error {
	cl, err := utils.NewClient(addr.String())
	if err != nil {
		return err
	}

	defer utils.CloseClient(cl)

	if err := utils.FillBenchmarkSourceMailbox(cl); err != nil {
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

	fetchCount := uint32(*fetchCountFlag)
	if fetchCount == 0 {
		fetchCount = messageCount / 2
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
}

func (*Fetch) TearDown(ctx context.Context, addr net.Addr) error {
	return nil
}

func (f *Fetch) Run(ctx context.Context, addr net.Addr) error {
	utils.RunParallelClients(addr, func(cl *client.Client, index uint) {
		var fetchFn func(*client.Client, *imap.SeqSet) error
		if *flags.UIDMode {
			fetchFn = func(cl *client.Client, set *imap.SeqSet) error {
				return utils.UIDFetchMessage(cl, set, imap.FetchAll)
			}
		} else {
			fetchFn = func(cl *client.Client, set *imap.SeqSet) error {
				return utils.FetchMessage(cl, set, imap.FetchAll)
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
