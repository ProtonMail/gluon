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
var copyAllFlag = flag.Bool("copy-all", false, "If set, perform a copy of the all messages.")

type Copy struct {
	seqSets    *ParallelSeqSet
	dstMailbox string
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

	messageCount := status.Messages

	if messageCount == 0 {
		return fmt.Errorf("mailbox '%v' has no messages", *flags.MailboxFlag)
	}

	copyCount := uint32(*copyCountFlag)
	if copyCount == 0 {
		copyCount = uint32(messageCount / 2)
	}

	seqSets, err := NewParallelSeqSet(copyCount,
		*flags.ParallelClientsFlag,
		*copyListFlag,
		*copyAllFlag,
		*flags.FlagRandomSeqSetIntervals)

	if err != nil {
		return err
	}

	c.seqSets = seqSets

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
		for _, v := range c.seqSets.Get(index) {
			if err := cl.Copy(v, c.dstMailbox); err != nil {
				panic(err)
			}
		}
	})

	return nil
}
