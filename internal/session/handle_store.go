package session

import (
	"context"
	"errors"

	"github.com/ProtonMail/gluon/internal/backend"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
)

func (s *Session) handleStore(ctx context.Context, tag string, cmd *proto.Store, mailbox *backend.Mailbox, ch chan response.Response) error {
	if cmd.GetAction().GetSilent() {
		ctx = backend.AsSilent(ctx)
	}

	flags, err := validateStoreFlags(cmd.GetFlags())
	if err != nil {
		return response.Bad(tag).WithError(err)
	}

	if err := mailbox.Store(ctx, cmd.GetSequenceSet(), cmd.GetAction().GetOperation(), flags); errors.Is(err, backend.ErrNoSuchMessage) {
		return response.Bad(tag).WithError(err)
	} else if err != nil {
		return err
	}

	if err := flush(ctx, mailbox, false, ch); err != nil {
		return err
	}

	var items []response.Item

	if mailbox.ExpungeIssued() {
		items = append(items, response.ItemExpungeIssued())
	}

	ch <- response.Ok(tag).
		WithItems(items...).
		WithMessage(okMessage(ctx))

	return nil
}
