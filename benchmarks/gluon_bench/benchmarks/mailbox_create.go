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
	if err != nil {
		return err
	}

	defer utils.CloseClient(cl)

	return utils.BuildMailbox(cl, "INBOX", 1000)
}
