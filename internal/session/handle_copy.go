package session

import (
	"context"
	"errors"

	"github.com/ProtonMail/gluon/imap/command"
	"github.com/ProtonMail/gluon/internal/contexts"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/ProtonMail/gluon/profiling"
	"github.com/ProtonMail/gluon/reporter"
)

func (s *Session) handleCopy(ctx context.Context, tag string, cmd *command.Copy, mailbox *state.Mailbox, ch chan response.Response) (response.Response, error) {
	if contexts.IsUID(ctx) {
		profiling.Start(ctx, profiling.CmdTypeUIDCopy)
		defer profiling.Stop(ctx, profiling.CmdTypeUIDCopy)
	} else {
		profiling.Start(ctx, profiling.CmdTypeCopy)
		defer profiling.Stop(ctx, profiling.CmdTypeCopy)
	}

	nameUTF8, err := s.decodeMailboxName(cmd.Mailbox)
	if err != nil {
		return nil, err
	}

	item, err := mailbox.Copy(ctx, cmd.SeqSet, nameUTF8)
	if errors.Is(err, state.ErrNoSuchMessage) {
		return response.Bad(tag).WithError(err), nil
	} else if errors.Is(err, state.ErrNoSuchMailbox) {
		return response.No(tag).WithError(err).WithItems(response.ItemTryCreate()), nil
	} else if err != nil {
		if shouldReportIMAPCommandError(err) {
			reporter.MessageWithContext(ctx,
				"Failed to copy messages",
				reporter.Context{"error": err, "mailbox_to": nameUTF8, "mailbox_from": mailbox.Name()},
			)
		}

		return nil, err
	}

	response := response.Ok(tag)

	if item != nil {
		response = response.WithItems(item)
	}

	return response.WithMessage(okMessage(ctx)), nil
}
