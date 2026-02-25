package session

import (
	"context"

	"github.com/ProtonMail/gluon/imap/command"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/profiling"
)

func (s *Session) handleGetQuota(ctx context.Context, tag string, cmd *command.GetQuota, ch chan response.Response) error {
	profiling.Start(ctx, profiling.CmdTypeGetQuota)
	defer profiling.Stop(ctx, profiling.CmdTypeGetQuota)

	root, err := s.state.GetQuota(ctx, cmd.Root)
	if err != nil {
		return err
	}

	ch <- response.Quota(root)

	ch <- response.Ok(tag).WithMessage("GETQUOTA")

	return nil
}
