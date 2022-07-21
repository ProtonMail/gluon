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

var fetchCountFlag = flag.Uint("fetch-count", 0, "Total number of messages to fetch during fetch benchmarks.")
var fetchListFlag = flag.String("fetch-list", "", "Use a list of predefined sequences to fetch rather than random generated.")
var fetchRandomRangesFlag = flag.Bool("fetch-random-ranges", false, "If not using a fetch list, use a random range rather than a random sequence.")
var fetchReadOnly = flag.Bool("fetch-read-only", false, "If set, perform fetches in read-only mode.")
var fetchAllFlag = flag.Bool("fetch-all", false, "If set, perform one fetch for all messages.")

type Fetch struct {
	messageCount uint32
	fetchCount   uint32
	fetchLists   [][]*imap.SeqSet
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

	status, err := cl.Status(*flags.MailboxFlag, []imap.StatusItem{imap.StatusMessages})
	if err != nil {
		return err
	}

	f.messageCount = status.Messages

	if f.messageCount == 0 {
		return fmt.Errorf("mailbox '%v' has no messages", *flags.MailboxFlag)
	}

	if *fetchCountFlag == 0 {
		f.fetchCount = f.messageCount / 2
	} else {
		f.fetchCount = uint32(*fetchCountFlag)
	}

	if len(*fetchListFlag) != 0 {
		list, err := utils.SequenceListFromFile(*fetchListFlag)
		if err != nil {
			return err
		}

		f.fetchLists = make([][]*imap.SeqSet, *flags.ParallelClientsFlag)
		for i := uint(0); i < *flags.ParallelClientsFlag; i++ {
			f.fetchLists[i] = list
		}
	} else if *fetchAllFlag {
		f.fetchLists = make([][]*imap.SeqSet, *flags.ParallelClientsFlag)
		for i := uint(0); i < *flags.ParallelClientsFlag; i++ {
			f.fetchLists[i] = []*imap.SeqSet{utils.NewSequenceSetAll()}
		}
	} else {
		f.fetchLists = make([][]*imap.SeqSet, *flags.ParallelClientsFlag)
		for i := uint(0); i < *flags.ParallelClientsFlag; i++ {
			list := make([]*imap.SeqSet, 0, f.fetchCount)
			for r := uint32(0); r < f.fetchCount; r++ {
				var seqSet *imap.SeqSet
				if !*fetchRandomRangesFlag {
					seqSet = utils.RandomSequenceSetNum(f.fetchCount)
				} else {
					seqSet = utils.RandomSequenceSetRange(f.fetchCount)
				}
				list = append(list, seqSet)
			}
			f.fetchLists[i] = list
		}
	}

	return nil
}

func (*Fetch) TearDown(ctx context.Context, addr net.Addr) error {
	return nil
}

func (f *Fetch) Run(ctx context.Context, addr net.Addr) error {
	utils.RunParallelClients(addr, func(cl *client.Client, index uint) {
		for _, v := range f.fetchLists[index] {
			if err := utils.FetchMessage(cl, v, imap.FetchAll); err != nil {
				panic(err)
			}
		}
	})

	return nil
}
