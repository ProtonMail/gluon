package session

import (
	"context"

	"github.com/ProtonMail/gluon/imap/command"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/profiling"
)

func (s *Session) handleGetQuotaRoot(ctx context.Context, tag string, cmd *command.GetQuotaRoot, ch chan response.Response) error {
	profiling.Start(ctx, profiling.CmdTypeGetQuotaRoot)
	defer profiling.Stop(ctx, profiling.CmdTypeGetQuotaRoot)

	rootNames, roots, err := s.state.GetQuotaRoot(ctx, cmd.Mailbox)
	if err != nil {
		return err
	}

	ch <- response.QuotaRoot(cmd.Mailbox, rootNames)

	for _, root := range roots {
		ch <- response.Quota(root)
	}

	ch <- response.Ok(tag).WithMessage("GETQUOTAROOT")

	return nil
}
