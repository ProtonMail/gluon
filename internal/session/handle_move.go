package session

import (
	"context"
	"errors"

	"github.com/ProtonMail/gluon/imap/command"
	"github.com/ProtonMail/gluon/internal/contexts"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/ProtonMail/gluon/profiling"
)

func (s *Session) handleMove(ctx context.Context, tag string, cmd *command.Move, mailbox *state.Mailbox, ch chan response.Response) (response.Response, error) {
	if contexts.IsUID(ctx) {
		profiling.Start(ctx, profiling.CmdTypeUIDMove)
		defer profiling.Stop(ctx, profiling.CmdTypeUIDMove)
	} else {
		profiling.Start(ctx, profiling.CmdTypeMove)
		defer profiling.Stop(ctx, profiling.CmdTypeMove)
	}

	nameUTF8, err := s.decodeMailboxName(cmd.Mailbox)
	if err != nil {
		return nil, err
	}

	if mailbox.ReadOnly() {
		return nil, ErrReadOnly
	}

	item, err := mailbox.Move(ctx, cmd.SeqSet, nameUTF8)
	if errors.Is(err, state.ErrNoSuchMessage) {
		return response.Bad(tag).WithError(err), nil
	} else if errors.Is(err, state.ErrNoSuchMailbox) {
		return response.No(tag).WithError(err).WithItems(response.ItemTryCreate()), nil
	} else if err != nil {
		return nil, err
	}

	if item != nil {
		ch <- response.Ok().WithItems(item)
	}

	if err := flush(ctx, mailbox, true, ch); err != nil {
		return nil, err
	}

	return response.Ok(tag).WithMessage(okMessage(ctx)), nil
}
