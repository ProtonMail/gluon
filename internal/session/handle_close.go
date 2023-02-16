package session

import (
	"context"

	"github.com/ProtonMail/gluon/imap/command"
	"github.com/ProtonMail/gluon/internal/contexts"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/ProtonMail/gluon/profiling"
)

func (s *Session) handleClose(ctx context.Context, tag string, _ *command.Close, mailbox *state.Mailbox, ch chan response.Response) (response.Response, error) {
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
