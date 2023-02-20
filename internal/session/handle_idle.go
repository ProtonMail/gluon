package session

import (
	"context"
	"time"

	"github.com/ProtonMail/gluon/imap/command"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/logging"
	"github.com/ProtonMail/gluon/profiling"
	"github.com/sirupsen/logrus"
)

// GOMSRV-86: What does it mean to do IDLE when you're not selected?
// GOMSRV-87: Should IDLE be stopped automatically when the context is cancelled?
func (s *Session) handleIdle(ctx context.Context, tag string, _ *command.Idle, cmdCh <-chan commandResult) error {
	profiling.Start(ctx, profiling.CmdTypeIdle)
	defer profiling.Stop(ctx, profiling.CmdTypeIdle)

	if s.state == nil {
		return ErrNotAuthenticated
	}

	return s.state.Idle(ctx, func(pending []response.Response, resCh chan response.Response) error {
		logging.GoAnnotated(ctx, func(ctx context.Context) {
			if s.idleBulkTime != 0 {
				sendResponsesInBulks(s, resCh, s.idleBulkTime)
			} else {
				for res := range resCh {
					if err := res.Send(s); err != nil {
						logrus.WithError(err).Error("Failed to send IDLE update")
					}
				}
			}
		}, logging.Labels{
			"Action":    "Sending IDLE updates",
			"SessionID": s.sessionID,
		})

		if err := response.Continuation().Send(s); err != nil {
			return err
		}

		for _, res := range pending {
			if err := res.Send(s); err != nil {
				return err
			}
		}

		var cmd commandResult

		for {
			select {
			case res, ok := <-cmdCh:
				if !ok {
					return nil
				}

				if res.err != nil {
					return res.err
				}
				cmd = res

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

			switch cmd.command.Payload.(type) {
			case *command.Done:
				return response.Ok(tag).WithMessage("IDLE").Send(s)

			default:
				return response.Bad(tag).Send(s)
			}
		}
	})
}

func sendMergedResponses(s *Session, buffer []response.Response) {
	for _, res := range response.Merge(buffer) {
		if err := res.Send(s); err != nil {
			logrus.WithError(err).Error("Failed to send IDLE update")
		}
	}
}

func sendResponsesInBulks(s *Session, resCh chan response.Response, idleBulkTime time.Duration) {
	buffer := []response.Response{}
	ticker := time.NewTicker(idleBulkTime)

	defer ticker.Stop()

	for {
		select {
		case res, ok := <-resCh:
			if !ok {
				sendMergedResponses(s, buffer)
				return
			}

			if res != nil {
				buffer = append(buffer, res)
				logrus.WithField("response", res).Trace("Buffered")
			}
		case <-ticker.C:
			sendMergedResponses(s, buffer)
			buffer = []response.Response{}
		}
	}
}
