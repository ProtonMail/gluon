package session

import (
	"context"

	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/logging"
	"github.com/ProtonMail/gluon/reporter"
)

type command struct {
	tag string
	cmd *proto.Command
	err error
}

func (s *Session) startCommandReader(ctx context.Context, del string) <-chan command {
	cmdCh := make(chan command)

	logging.GoAnnotate(ctx, func(ctx context.Context) {
		defer close(cmdCh)

		for {
			tag, cmd, err := s.readCommand(del)
			if err != nil {
				reporter.MessageWithContext(ctx,
					"Failed to parse imap command",
					reporter.Context{"error": err},
				)
			}

			if err == nil && cmd.GetStartTLS() != nil {
				// TLS needs to be handled here to ensure that next command read is over the TLS connection.
				if startTLSErr := s.handleStartTLS(tag, cmd.GetStartTLS()); startTLSErr != nil {
					cmd = nil
					err = startTLSErr
				} else {
					continue
				}
			}

			select {
			case cmdCh <- command{tag: tag, cmd: cmd, err: err}:
				// ...

			case <-ctx.Done():
				return
			}
		}
	}, map[string]any{
		"Action":    "Reading IMAP commands",
		"SessionID": s.sessionID,
	})

	return cmdCh
}

func (s *Session) readCommand(del string) (string, *proto.Command, error) {
	line, literals, err := s.liner.Read(func() error { return response.Continuation().Send(s) })
	if err != nil {
		return "", nil, err
	}

	s.logIncoming(string(line))

	tag, cmd, err := parse(line, literals, del)
	if err != nil {
		return tag, cmd, err
	}

	return tag, cmd, nil
}
