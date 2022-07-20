package benchmarks

import (
	"context"
	"net"

	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/utils"
)

type MailboxCreate struct{}

func (b *MailboxCreate) Name() string {
	return "mailbox-create"
}

func (*MailboxCreate) Setup(ctx context.Context, addr net.Addr) error {
	return nil
}

func (*MailboxCreate) TearDown(ctx context.Context, addr net.Addr) error {
	return nil
}

func (*MailboxCreate) Run(ctx context.Context, addr net.Addr) error {
	cl, err := utils.NewClient(addr.String())
	defer utils.CloseClient(cl)

	if err != nil {
		return err
	}

	return utils.BuildMailbox(cl, "INBOX", 1000)
}
