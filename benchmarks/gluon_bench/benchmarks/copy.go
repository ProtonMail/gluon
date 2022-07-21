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
	"github.com/google/uuid"
)

var copyCountFlag = flag.Uint("copy-count", 0, "Total number of messages to copy during copy benchmarks.")
var copyListFlag = flag.String("copy-list", "", "Use a list of predefined sequences to copy rather than random generated.")
var copyRandomRangesFlag = flag.Bool("copy-random-ranges", false, "If not using a copy list, use a random range rather than a random sequence.")
var copyAllFlag = flag.Bool("copy-all", false, "If set, perform a copy of the all messages.")

type Copy struct {
	messageCount uint32
	copyCount    uint32
	copyLists    [][]*imap.SeqSet
	dstMailbox   string
}

func NewCopy() *Copy {
	return &Copy{
		dstMailbox: uuid.NewString(),
	}
}

func (*Copy) Name() string {
	return "copy"
}

func (c *Copy) Setup(ctx context.Context, addr net.Addr) error {
	cl, err := utils.NewClient(addr.String())
	if err != nil {
		return err
	}

	defer utils.CloseClient(cl)

	if err := utils.FillBenchmarkSourceMailbox(cl); err != nil {
		return err
	}

	//Delete mailbox if it exists
	if err := cl.Delete(c.dstMailbox); err != nil {
		// ignore error
	}

	if err := cl.Create(c.dstMailbox); err != nil {
		return err
	}

	status, err := cl.Status(*flags.MailboxFlag, []imap.StatusItem{imap.StatusMessages})
	if err != nil {
		return err
	}

	c.messageCount = status.Messages

	if c.messageCount == 0 {
		return fmt.Errorf("mailbox '%v' has no messages", *flags.MailboxFlag)
	}

	if *copyCountFlag == 0 {
		c.copyCount = c.messageCount / 2
	} else {
		c.copyCount = uint32(*copyCountFlag)
	}

	if len(*copyListFlag) != 0 {
		list, err := utils.SequenceListFromFile(*copyListFlag)
		if err != nil {
			return err
		}

		c.copyLists = make([][]*imap.SeqSet, *flags.ParallelClientsFlag)
		for i := uint(0); i < *flags.ParallelClientsFlag; i++ {
			c.copyLists[i] = list
		}
	} else if *copyAllFlag {
		c.copyLists = make([][]*imap.SeqSet, *flags.ParallelClientsFlag)
		for i := uint(0); i < *flags.ParallelClientsFlag; i++ {
			c.copyLists[i] = []*imap.SeqSet{utils.NewSequenceSetAll()}
		}
	} else {
		c.copyLists = make([][]*imap.SeqSet, *flags.ParallelClientsFlag)
		for i := uint(0); i < *flags.ParallelClientsFlag; i++ {
			list := make([]*imap.SeqSet, 0, c.copyCount)
			for r := uint32(0); r < c.copyCount; r++ {
				var seqSet *imap.SeqSet
				if !*copyRandomRangesFlag {
					seqSet = utils.RandomSequenceSetNum(c.copyCount)
				} else {
					seqSet = utils.RandomSequenceSetRange(c.copyCount)
				}
				list = append(list, seqSet)
			}
			c.copyLists[i] = list
		}
	}

	return nil
}

func (c *Copy) TearDown(ctx context.Context, addr net.Addr) error {
	cl, err := utils.NewClient(addr.String())
	if err != nil {
		return err
	}

	defer utils.CloseClient(cl)

	if err := cl.Delete(c.dstMailbox); err != nil {
		return err
	}

	return nil
}

func (c *Copy) Run(ctx context.Context, addr net.Addr) error {
	utils.RunParallelClients(addr, func(cl *client.Client, index uint) {
		for _, v := range c.copyLists[index] {
			if err := cl.Copy(v, c.dstMailbox); err != nil {
				panic(err)
			}
		}
	})

	return nil
}
