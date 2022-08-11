package session

import (
	"context"
	"errors"

	"github.com/ProtonMail/gluon/internal/backend"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/emersion/go-imap/utf7"
)

func (s *Session) handleCopy(ctx context.Context, tag string, cmd *proto.Copy, mailbox *backend.Mailbox, ch chan response.Response) error {
	nameUTF8, err := utf7.Encoding.NewDecoder().String(cmd.GetMailbox())
	if err != nil {
		return err
	}

	item, err := mailbox.Copy(ctx, cmd.GetSequenceSet(), nameUTF8)
	if errors.Is(err, backend.ErrNoSuchMessage) {
		return response.Bad(tag).WithError(err)
	} else if errors.Is(err, backend.ErrNoSuchMailbox) {
		return response.No(tag).WithError(err).WithItems(response.ItemTryCreate())
	} else if err != nil {
		return err
	}

	response := response.Ok(tag)

	if item != nil {
		response = response.WithItems(item)
	}

	ch <- response.WithMessage(okMessage(ctx))

	return nil
}
