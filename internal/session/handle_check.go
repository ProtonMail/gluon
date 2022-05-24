package session

import (
	"context"

	"github.com/ProtonMail/gluon/internal/backend"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
)

func (s *Session) handleCheck(ctx context.Context, tag string, cmd *proto.Check, mailbox *backend.Mailbox, ch chan response.Response) error {
	if err := flush(ctx, mailbox, true, ch); err != nil {
		return err
	}

	ch <- response.Ok(tag)

	return nil
}
