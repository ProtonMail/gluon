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

func (s *Session) handleCopy(ctx context.Context, tag string, cmd *proto.Copy, mailbox *state.Mailbox, ch chan response.Response) (response.Response, error) {
	nameUTF8, err := utf7.Encoding.NewDecoder().String(cmd.GetMailbox())
	if err != nil {
		return nil, err
	}

	item, err := mailbox.Copy(ctx, cmd.GetSequenceSet(), nameUTF8)
	if errors.Is(err, state.ErrNoSuchMessage) {
		return response.Bad(tag).WithError(err), nil
	} else if errors.Is(err, state.ErrNoSuchMailbox) {
		return response.No(tag).WithError(err).WithItems(response.ItemTryCreate()), nil
	} else if err != nil {
		reporter.MessageWithContext(ctx,
			"Failed to copy messages",
			reporter.Context{"error": err},
		)

		return nil, err
	}

	response := response.Ok(tag)

	if item != nil {
		response = response.WithItems(item)
	}

	return response.WithMessage(okMessage(ctx)), nil
}
