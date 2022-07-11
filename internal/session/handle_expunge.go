package session

import (
	"context"

	"github.com/ProtonMail/gluon/internal/backend"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
)

func (s *Session) handleExpunge(ctx context.Context, tag string, cmd *proto.Expunge, mailbox *backend.Mailbox, ch chan response.Response) error {
	if mailbox.ReadOnly() {
		return ErrReadOnly
	}

	if err := mailbox.Expunge(ctx, nil); err != nil {
		return err
	}

	if err := flush(ctx, mailbox, true, ch); err != nil {
		return err
	}

	ch <- response.Ok(tag).WithMessage("EXPUNGE")

	return nil
}

func (s *Session) handleUIDExpunge(ctx context.Context, tag string, cmd *proto.UIDExpunge, mailbox *backend.Mailbox, ch chan response.Response) error {
	ctx = backend.AsUID(ctx)

	if mailbox.ReadOnly() {
		return ErrReadOnly
	}

	if err := mailbox.Expunge(ctx, cmd.GetSequenceSet()); err != nil {
		return err
	}

	if err := flush(ctx, mailbox, true, ch); err != nil {
		return err
	}

	ch <- response.Ok(tag).WithMessage("EXPUNGE")

	return nil
}
