package benchmarks

import (
	"context"
	"fmt"

	"github.com/ProtonMail/gluon"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/utils"
)

func BenchmarkMailboxCreate(ctx context.Context, server *gluon.Server, addr string) {
	cl, err := utils.NewClient(addr)
	defer utils.CloseClient(cl)

	if err != nil {
		panic("Failed to connect o server")
	}

	if err := utils.BuildMailbox(cl, "INBOX", 1000); err != nil {
		panic(fmt.Sprintf("Benchmark Mailbox Create - Failed to create mailbox: %v", err))
	}
}
