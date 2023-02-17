package session

import (
	"context"

	"github.com/ProtonMail/gluon/imap/command"
	"github.com/ProtonMail/gluon/internal/contexts"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/ProtonMail/gluon/profiling"
)

func (s *Session) handleExpunge(ctx context.Context, tag string, cmd *command.Expunge, mailbox *state.Mailbox, ch chan response.Response) (response.Response, error) {
	profiling.Start(ctx, profiling.CmdTypeExpunge)
	defer profiling.Stop(ctx, profiling.CmdTypeExpunge)

	if mailbox.ReadOnly() {
		return nil, ErrReadOnly
	}

	if err := mailbox.Expunge(ctx, nil); err != nil {
		return nil, err
	}

	if err := flush(ctx, mailbox, true, ch); err != nil {
		return nil, err
	}

	return response.Ok(tag).WithMessage("EXPUNGE"), nil
}

func (s *Session) handleUIDExpunge(ctx context.Context, tag string, cmd *command.UIDExpunge, mailbox *state.Mailbox, ch chan response.Response) (response.Response, error) {
	profiling.Start(ctx, profiling.CmdTypeExpunge)
	defer profiling.Stop(ctx, profiling.CmdTypeExpunge)

	ctx = contexts.AsUID(ctx)

	if mailbox.ReadOnly() {
		return nil, ErrReadOnly
	}

	if err := mailbox.Expunge(ctx, cmd.SeqSet); err != nil {
		return nil, err
	}

	if err := flush(ctx, mailbox, true, ch); err != nil {
		return nil, err
	}

	return response.Ok(tag).WithMessage("EXPUNGE"), nil
}
