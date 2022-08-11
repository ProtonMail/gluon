package session

import (
	"context"
	"errors"
	"io"
	"runtime/pprof"
	"strconv"

	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
)

type command struct {
	tag string
	cmd *proto.Command
}

func (s *Session) getCommandCh(ctx context.Context, del string) <-chan command {
	cmdCh := make(chan command)

	go func() {
		labels := pprof.Labels("go", "CommandReader", "SessionID", strconv.Itoa(s.sessionID))
		pprof.Do(ctx, labels, func(_ context.Context) {
			defer close(cmdCh)
			for {
				tag, cmd, err := s.readCommand(del)
				if err != nil {
					if errors.Is(err, io.EOF) {
						return
					} else if err := response.Bad(tag).WithError(err).Send(s); err != nil {
						return
					}

					continue
				}

				switch {
				case cmd.GetStartTLS() != nil:
					if err := s.handleStartTLS(tag, cmd.GetStartTLS()); err != nil {
						if err := response.Bad(tag).WithError(err).Send(s); err != nil {
							return
						}

						continue
					}

				default:
					cmdCh <- command{tag: tag, cmd: cmd}
				}
			}
		})
	}()

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
