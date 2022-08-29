package session

import (
	"context"
	"runtime/pprof"
	"strconv"

	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/sirupsen/logrus"
)

// GOMSRV-86: What does it mean to do IDLE when you're not selected?
// GOMSRV-87: Should IDLE be stopped automatically when the context is cancelled?
func (s *Session) handleIdle(ctx context.Context, tag string, cmd *proto.Idle, cmdCh <-chan command) error {
	if s.state == nil {
		return ErrNotAuthenticated
	}

	return s.state.Idle(ctx, func(pending []response.Response, resCh chan response.Response) error {
		go func() {
			labels := pprof.Labels("go", "Idle", "SessionID", strconv.Itoa(s.sessionID))
			pprof.Do(ctx, labels, func(_ context.Context) {
				for res := range resCh {
					if err := res.Send(s); err != nil {
						logrus.WithError(err).Error("Failed to send IDLE update")
					}
				}
			})
		}()

		if err := response.Continuation().Send(s); err != nil {
			return err
		}

		for _, res := range pending {
			if err := res.Send(s); err != nil {
				return err
			}
		}

		var cmd *proto.Command

		for {
			select {
			case res, ok := <-cmdCh:
				if !ok {
					return nil
				}

				cmd = res.cmd

			case <-s.state.Done():
				return nil

			case stateUpdate := <-s.state.GetStateUpdatesCh():
				if err := s.state.ApplyUpdate(ctx, stateUpdate); err != nil {
					logrus.WithError(err).Error("Failed to apply state update during idle")
				}
				continue

			case <-ctx.Done():
				return ctx.Err()
			}

			switch {
			case cmd.GetDone() != nil:
				return response.Ok(tag).WithMessage("IDLE").Send(s)

			default:
				return response.Bad(tag).Send(s)
			}
		}
	})
}
