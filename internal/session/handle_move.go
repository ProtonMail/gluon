package session

import (
	"context"
	"errors"

	"github.com/ProtonMail/gluon/internal/contexts"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/ProtonMail/gluon/profiling"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/emersion/go-imap/utf7"
)

func (s *Session) handleMove(ctx context.Context, tag string, cmd *proto.Move, mailbox *state.Mailbox, ch chan response.Response) (response.Response, error) {
	if contexts.IsUID(ctx) {
		profiling.Start(ctx, profiling.CmdTypeUIDMove)
		defer profiling.Stop(ctx, profiling.CmdTypeUIDMove)
	} else {
		profiling.Start(ctx, profiling.CmdTypeMove)
		defer profiling.Stop(ctx, profiling.CmdTypeMove)
	}

	nameUTF8, err := utf7.Encoding.NewDecoder().String(cmd.GetMailbox())
	if err != nil {
		return nil, err
	}

	item, err := mailbox.Move(ctx, cmd.GetSequenceSet(), nameUTF8)
	if errors.Is(err, state.ErrNoSuchMessage) {
		return response.Bad(tag).WithError(err), nil
	} else if errors.Is(err, state.ErrNoSuchMailbox) {
		return response.No(tag).WithError(err).WithItems(response.ItemTryCreate()), nil
	} else if err != nil {
		if shouldReportIMAPCommandError(err) {
			reporter.MessageWithContext(ctx,
				"Failed to move messages from mailbox",
				reporter.Context{"error": err, "mailbox_to": nameUTF8, "mailbox_from": mailbox.Name()},
			)
		}

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
