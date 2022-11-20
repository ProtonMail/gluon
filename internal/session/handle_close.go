package session

import (
	"context"

	"github.com/ProtonMail/gluon/internal/contexts"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/ProtonMail/gluon/profiling"
)

// TODO(REFACTOR): No EXPUNGE responses are sent -- what about to other sessions also selected in this mailbox?
func (s *Session) handleClose(ctx context.Context, tag string, cmd *proto.Close, mailbox *state.Mailbox, ch chan response.Response) (response.Response, error) {
	profiling.Start(ctx, profiling.CmdTypeClose)
	defer profiling.Stop(ctx, profiling.CmdTypeClose)

	ctx = contexts.AsClose(ctx)

	if !mailbox.ReadOnly() {
		if err := mailbox.Expunge(ctx, nil); err != nil {
			return nil, err
		}
	}

	if err := flush(ctx, mailbox, true, ch); err != nil {
		return nil, err
	}

	if err := mailbox.Close(ctx); err != nil {
		return nil, err
	}

	return response.Ok(tag).WithMessage("CLOSE"), nil
}
