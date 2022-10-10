package session

import (
	"context"
	"github.com/ProtonMail/gluon/internal/parser"
	"runtime/pprof"
	"strconv"

	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/reporter"
)

type command struct {
	tag string
	cmd *proto.Command
	err error
}

func (s *Session) startCommandReader(ctx context.Context, del rune) <-chan command {
	cmdCh := make(chan command)

	go func() {
		labels := pprof.Labels("go", "CommandReader", "SessionID", strconv.Itoa(s.sessionID))
		pprof.Do(ctx, labels, func(_ context.Context) {
			defer close(cmdCh)

			imapParser := parser.NewIMAPParser()
			defer imapParser.Close()

			for {
				tag, cmd, err := s.readCommand(imapParser, del)
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
		})
	}()

	return cmdCh
}

func (s *Session) readCommand(parser *parser.IMAPParser, del rune) (string, *proto.Command, error) {
	line, err := s.liner.Read(func() error { return response.Continuation().Send(s) })
	if err != nil {
		return "", nil, err
	}

	s.logIncoming(line)

	return parser.Parse(line, del)
}
