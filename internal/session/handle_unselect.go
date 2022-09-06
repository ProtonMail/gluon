package session

import (
	"context"

	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
)

func (s *Session) handleUnselect(ctx context.Context, tag string, cmd *proto.Unselect, mailbox *state.Mailbox, _ chan response.Response) (response.Response, error) {
	if err := mailbox.Close(ctx); err != nil {
		return nil, err
	}

	return response.Ok(tag).WithMessage("UNSELECT"), nil
}
