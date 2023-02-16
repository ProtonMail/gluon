package session

import (
	"context"

	"github.com/ProtonMail/gluon/imap/command"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/ProtonMail/gluon/profiling"
)

func (s *Session) handleNoop(ctx context.Context, tag string, _ *command.Noop, ch chan response.Response) error {
	profiling.Start(ctx, profiling.CmdTypeNoop)
	defer profiling.Stop(ctx, profiling.CmdTypeNoop)

	if (s.state != nil) && s.state.IsSelected() {
		if err := s.state.Selected(ctx, func(mailbox *state.Mailbox) error {
			return flush(ctx, mailbox, true, ch)
		}); err != nil {
			return err
		}
	}

	ch <- response.Ok(tag).WithMessage(okMessage(ctx))

	return nil
}
