package session

import (
	"context"

	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/sirupsen/logrus"
)

// GOMSRV-86: What does it mean to do IDLE when you're not selected?
// GOMSRV-87: Should IDLE be stopped automatically when the context is cancelled?
func (s *Session) handleIdle(ctx context.Context, tag string, cmd *proto.Idle) error {
	if s.state == nil {
		return ErrNotAuthenticated
	}

	return s.state.Idle(ctx, func(pending []response.Response, resCh chan response.Response) error {
		go func() {
			for res := range resCh {
				if err := res.Send(s); err != nil {
					logrus.WithError(err).Error("Failed to send IDLE update")
				}
			}
		}()

		if err := response.Continuation().Send(s); err != nil {
			return err
		}

		for _, res := range pending {
			if err := res.Send(s); err != nil {
				return err
			}
		}

		for {
			_, cmd, err := s.readCommand()
			if err != nil {
				return err
			}

			switch {
			case cmd.GetDone() != nil:
				return response.Ok(tag).Send(s)

			default:
				return response.Bad(tag).Send(s)
			}
		}
	})
}
