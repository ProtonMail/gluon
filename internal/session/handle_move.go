package session

import (
	"context"
	"errors"
	"strings"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/backend"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/emersion/go-imap/utf7"
)

func (s *Session) handleMove(ctx context.Context, tag string, cmd *proto.Move, mailbox *backend.Mailbox, ch chan response.Response) error {
	nameUTF8, err := utf7.Encoding.NewDecoder().String(cmd.GetMailbox())
	if err != nil {
		return err
	}

	if strings.EqualFold(nameUTF8, imap.Inbox) {
		nameUTF8 = imap.Inbox
	}

	item, err := mailbox.Move(ctx, cmd.GetSequenceSet(), nameUTF8)
	if errors.Is(err, backend.ErrNoSuchMessage) {
		return response.Bad(tag).WithError(err)
	} else if errors.Is(err, backend.ErrNoSuchMailbox) {
		return response.No(tag).WithError(err).WithItems(response.ItemTryCreate())
	} else if err != nil {
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
