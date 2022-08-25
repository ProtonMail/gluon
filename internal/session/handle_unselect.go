package session

import (
	"context"

	"github.com/ProtonMail/gluon/internal/backend"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
)

func (s *Session) handleUnselect(ctx context.Context, tag string, cmd *proto.Unselect, mailbox *backend.Mailbox, ch chan response.Response) error {
	if err := mailbox.Close(ctx); err != nil {
		return err
	}

	ch <- response.Ok(tag).WithMessage("UNSELECT")

	return nil
}
