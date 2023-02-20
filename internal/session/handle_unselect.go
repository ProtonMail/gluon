package session

import (
	"context"

	"github.com/ProtonMail/gluon/imap/command"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/ProtonMail/gluon/profiling"
)

func (s *Session) handleUnselect(ctx context.Context, tag string, _ *command.Unselect, mailbox *state.Mailbox, _ chan response.Response) (response.Response, error) {
	profiling.Start(ctx, profiling.CmdTypeUnselect)
	defer profiling.Stop(ctx, profiling.CmdTypeUnselect)

	if err := mailbox.Close(ctx); err != nil {
		return nil, err
	}

	return response.Ok(tag).WithMessage("UNSELECT"), nil
}
