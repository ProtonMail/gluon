package session

import (
	"context"

	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/ProtonMail/gluon/profiling"
)

func (s *Session) handleCheck(ctx context.Context, tag string, cmd *proto.Check, mailbox *state.Mailbox, ch chan response.Response) (response.Response, error) {
	profiling.Start(ctx, profiling.CmdTypeCheck)
	defer profiling.Stop(ctx, profiling.CmdTypeCheck)

	if err := flush(ctx, mailbox, true, ch); err != nil {
		return nil, err
	}

	return response.Ok(tag).WithMessage("CHECK"), nil
}
