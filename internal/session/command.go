package session

import (
	"errors"
	"io"

	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
)

type command struct {
	tag string
	cmd *proto.Command
}

func (s *Session) getCommandCh() <-chan command {
	cmdCh := make(chan command)

	go func() {
		defer close(cmdCh)

		for {
			tag, cmd, err := s.readCommand()
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
	}()

	return cmdCh
}

func (s *Session) readCommand() (string, *proto.Command, error) {
	line, literals, err := s.liner.Read(func() error { return response.Continuation().Send(s) })
	if err != nil {
		return "", nil, err
	}

	s.logIncoming(string(line))

	tag, cmd, err := parse(line, literals)
	if err != nil {
		return tag, cmd, err
	}

	return tag, cmd, nil
}
