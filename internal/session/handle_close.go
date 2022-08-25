package session

import (
	"context"

	"github.com/ProtonMail/gluon/internal/backend"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
)

// TODO(REFACTOR): No EXPUNGE responses are sent -- what about to other sessions also selected in this mailbox?
func (s *Session) handleClose(ctx context.Context, tag string, cmd *proto.Close, mailbox *backend.Mailbox, ch chan response.Response) error {
	ctx = backend.AsClose(ctx)

	if !mailbox.ReadOnly() {
		if err := mailbox.Expunge(ctx, nil); err != nil {
			return err
		}
	}

	if err := flush(ctx, mailbox, true, ch); err != nil {
		return err
	}

	if err := mailbox.Close(ctx); err != nil {
		return err
	}

	ch <- response.Ok(tag).WithMessage("CLOSE")

	return nil
}
