package session

import (
	"context"
	"encoding/base64"
	"fmt"
	"unicode/utf8"

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

			// If the input is not valid UTF-8, we attempt to handle it as an upgrade to TLS.
			// If that fails, we report the error to sentry and return.
			if !utf8.Valid(line) {
				if err := s.handleUpgrade(line); err != nil {
					reporter.MessageWithContext(ctx,
						"Received invalid UTF-8 command",
						reporter.Context{"line (base 64)": base64.StdEncoding.EncodeToString(line)},
					)

					return
				}

				continue
			}

			tag, cmd, err := parse(line, literals, del)
			if err != nil {
				reporter.MessageWithContext(ctx,
					"Failed to parse IMAP command",
					reporter.Context{"error": err},
				)
			} else if cmd.GetStartTLS() != nil {
				// TLS needs to be handled here to ensure that next command read is over the TLS connection.
				if err = s.handleStartTLS(tag); err != nil {
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
