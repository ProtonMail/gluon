package benchmarks

import (
	"context"
	"flag"
	"fmt"
	"net"

	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/flags"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/utils"
)

var messageCountFlag = flag.Uint("mailbox-create-message-count", 1000, "Number of random messages to create in the inbox.")

type MailboxCreate struct{}

func (b *MailboxCreate) Name() string {
	return "mailbox-create"
}

func (*MailboxCreate) Setup(ctx context.Context, addr net.Addr) error {
	if *messageCountFlag == 0 {
		return fmt.Errorf("invalid message count")
	}

	return nil
}

func (*MailboxCreate) TearDown(ctx context.Context, addr net.Addr) error {
	return nil
}

func (*MailboxCreate) Run(ctx context.Context, addr net.Addr) error {
	cl, err := utils.NewClient(addr.String())
	if err != nil {
		return err
	}

	defer utils.CloseClient(cl)

	return utils.BuildMailbox(cl, *flags.MailboxFlag, int(*messageCountFlag))
}
