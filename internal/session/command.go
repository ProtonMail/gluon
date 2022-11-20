package session

import (
	"context"
	"fmt"

	"github.com/bradenaw/juniper/xslices"
	"golang.org/x/exp/maps"

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

	logging.GoAnnotated(ctx, func(ctx context.Context) {
		defer close(cmdCh)

		for {
			line, literals, err := s.liner.Read(func() error { return response.Continuation().Send(s) })
			if err != nil {
				return
			}

			s.logIncoming(string(line), xslices.Map(maps.Keys(literals), func(k string) string {
				return fmt.Sprintf("%v: '%s'", k, literals[k])
			})...)

			tag, cmd, err := parse(line, literals, del)
			if err != nil {
				reporter.MessageWithContext(ctx,
					"Failed to parse imap command",
					reporter.Context{"error": err},
				)
			} else if cmd.GetStartTLS() != nil {
				// TLS needs to be handled here to ensure that next command read is over the TLS connection.
				if err = s.handleStartTLS(tag, cmd.GetStartTLS()); err != nil {
					return
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
	}, logging.Labels{
		"Action":    "Reading commands",
		"SessionID": s.sessionID,
	})

	return cmdCh
}
