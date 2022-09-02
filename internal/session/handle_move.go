package session

import (
	"context"
	"errors"

	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/emersion/go-imap/utf7"
)

func (s *Session) handleMove(ctx context.Context, tag string, cmd *proto.Move, mailbox *state.Mailbox, ch chan response.Response) error {
	nameUTF8, err := utf7.Encoding.NewDecoder().String(cmd.GetMailbox())
	if err != nil {
		return err
	}

	item, err := mailbox.Move(ctx, cmd.GetSequenceSet(), nameUTF8)
	if errors.Is(err, state.ErrNoSuchMessage) {
		return response.Bad(tag).WithError(err)
	} else if errors.Is(err, state.ErrNoSuchMailbox) {
		return response.No(tag).WithError(err).WithItems(response.ItemTryCreate())
	} else if err != nil {
		reporter.MessageWithContext(ctx,
			"Failed to move messages from mailbox",
			reporter.Context{"error": err},
		)

		return err
	}

	if item != nil {
		ch <- response.Ok().WithItems(item)
	}

	if err := flush(ctx, mailbox, true, ch); err != nil {
		return err
	}

	ch <- response.Ok(tag).WithMessage(okMessage(ctx))

	return nil
}
