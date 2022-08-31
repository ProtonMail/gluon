package session

import (
	"context"

	context2 "github.com/ProtonMail/gluon/internal/contexts"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
)

// TODO(REFACTOR): No EXPUNGE responses are sent -- what about to other sessions also selected in this mailbox?
func (s *Session) handleClose(ctx context.Context, tag string, cmd *proto.Close, mailbox *state.Mailbox, ch chan response.Response) error {
	ctx = context2.AsClose(ctx)

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
