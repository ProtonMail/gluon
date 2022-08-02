package benchmarks

import (
	"context"
	"flag"
	"fmt"
	"net"

	"github.com/bradenaw/juniper/xslices"

	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/flags"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/utils"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/google/uuid"
)

var (
	moveListFlag        = flag.String("move-list", "", "Use a list of predefined sequences to move rather than random generated.")
	moveAllFlag         = flag.Bool("move-all", false, "If set, perform a move of the all messages.")
	moveIntoSameDstFlag = flag.Bool("move-into-same-dst", false, "If set, rather than moving each unique mailbox into separate unique mailboxes, move all messages into one common destination mailbox.")
)

type Move struct {
	seqSets       *ParallelSeqSet
	messageCounts []uint32
	srcMailboxes  []string
	dstMailboxes  []string
}

func NewMove() *Move {
	return &Move{}
}

func (*Move) Name() string {
	return "move"
}

func (m *Move) Setup(ctx context.Context, addr net.Addr) error {
	if *flags.FillSourceMailbox == 0 {
		return fmt.Errorf("move benchmark requires a message count > 0")
	}

	cl, err := utils.NewClient(addr.String())
	if err != nil {
		return err
	}

	defer utils.CloseClient(cl)

	srcMailboxes := make([]string, 0, *flags.ParallelClients)
	dstMailboxes := make([]string, 0, *flags.ParallelClients)

	for i := uint(0); i < *flags.ParallelClients; i++ {
		srcMailboxes = append(srcMailboxes, uuid.NewString())
	}

	if *moveIntoSameDstFlag {
		dstMailboxes = []string{uuid.NewString()}
	} else {
		for i := uint(0); i < *flags.ParallelClients; i++ {
			dstMailboxes = append(dstMailboxes, uuid.NewString())
		}
	}

	m.srcMailboxes = srcMailboxes
	m.dstMailboxes = dstMailboxes

	// Delete mailboxes if they exist
	for _, v := range srcMailboxes {
		if err := cl.Delete(v); err != nil {
			// ignore errors
		}
	}

	for _, v := range dstMailboxes {
		if err := cl.Delete(v); err != nil {
			// ignore errors
		}
	}

	// Create mailboxes
	for _, v := range srcMailboxes {
		if err := cl.Create(v); err != nil {
			return err
		}
	}

	for _, v := range dstMailboxes {
		if err := cl.Create(v); err != nil {
			return err
		}
	}

	// Fill srcMailboxes
	for _, v := range srcMailboxes {
		if err := utils.BuildMailbox(cl, v, int(*flags.FillSourceMailbox)); err != nil {
			return err
		}
	}

	seqSets, err := NewParallelSeqSet(uint32(*flags.FillSourceMailbox),
		*flags.ParallelClients,
		*moveListFlag,
		*moveAllFlag,
		*flags.RandomSeqSetIntervals,
		true,
		*flags.UIDMode)

	if err != nil {
		return err
	}

	m.seqSets = seqSets

	return nil
}

func (m *Move) TearDown(ctx context.Context, addr net.Addr) error {
	cl, err := utils.NewClient(addr.String())
	if err != nil {
		return err
	}

	defer utils.CloseClient(cl)

	for _, v := range m.srcMailboxes {
		if err := cl.Delete(v); err != nil {
			return err
		}
	}

	for _, v := range m.dstMailboxes {
		if err := cl.Delete(v); err != nil {
			return err
		}
	}

	return nil
}

func (m *Move) Run(ctx context.Context, addr net.Addr) error {
	mboxInfos := xslices.Map(m.srcMailboxes, func(name string) utils.MailboxInfo {
		return utils.MailboxInfo{
			Name:     name,
			ReadOnly: true,
		}
	})

	utils.RunParallelClientsWithMailboxes(addr, mboxInfos, func(cl *client.Client, index uint) {
		var moveFn func(*client.Client, *imap.SeqSet, string) error
		if *flags.UIDMode {
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
