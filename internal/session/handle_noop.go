package session

import (
	"context"

	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
)

func (s *Session) handleNoop(ctx context.Context, tag string, cmd *proto.Noop, ch chan response.Response) error {
	if (s.state != nil) && s.state.IsSelected() {
		if err := s.state.Selected(ctx, func(mailbox *state.Mailbox) error {
			return flush(ctx, mailbox, true, ch)
		}); err != nil {
			return err
		}
	}

	ch <- response.Ok(tag).WithMessage(okMessage(ctx))

	return nil
}
