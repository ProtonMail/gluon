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
	storeCountFlag  = flag.Uint("store-count", 0, "Total number of messages to store during store benchmarks.")
	storeListFlag   = flag.String("store-list", "", "Use a list of predefined sequences to store rather than random generated.")
	storeSilentFlag = flag.Bool("store-silent", false, "When set to true, request silent updates that do not produce any responses")
)

type Store struct {
	seqSets *ParallelSeqSet
}

func NewStore() *Store {
	return &Store{}
}

func (*Store) Name() string {
	return "store"
}

func (s *Store) Setup(ctx context.Context, addr net.Addr) error {
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

	s.seqSets = seqSets

	return nil
}

func (*Store) TearDown(ctx context.Context, addr net.Addr) error {
	return nil
}

func (s *Store) Run(ctx context.Context, addr net.Addr) error {
	items := []string{"FLAGS", "-FLAGS", "+FLAGS"}
	flagList := []string{imap.DeletedFlag, imap.SeenFlag, imap.AnsweredFlag, imap.FlaggedFlag}

	utils.RunParallelClients(addr, false, func(cl *client.Client, index uint) {
		var storeFn func(*client.Client, *imap.SeqSet, int) error
		if *flags.UIDMode {
			storeFn = func(cl *client.Client, set *imap.SeqSet, index int) error {
				return utils.UIDStore(cl, set, items[index%len(items)], *storeSilentFlag, flagList[index%len(flagList)])
			}
		} else {
			storeFn = func(cl *client.Client, set *imap.SeqSet, index int) error {
				return utils.Store(cl, set, items[index%len(items)], *storeSilentFlag, flagList[index%len(flagList)])
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
