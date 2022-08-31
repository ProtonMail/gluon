package session

import (
	"context"

	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
)

func (s *Session) handleUnselect(ctx context.Context, tag string, cmd *proto.Unselect, mailbox *state.Mailbox, ch chan response.Response) error {
	if err := mailbox.Close(ctx); err != nil {
		return err
	}

	ch <- response.Ok(tag).WithMessage("UNSELECT")

	return nil
}
