package session

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"unicode/utf8"

	"github.com/bradenaw/juniper/xslices"
	"github.com/sirupsen/logrus"
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

		tlsHeaders := [][]byte{
			{0x16, 0x03, 0x01}, // 1.0
			{0x16, 0x03, 0x02}, // 1.1
			{0x16, 0x03, 0x03}, // 1.2
			{0x16, 0x03, 0x04}, // 1.3
			{0x16, 0x00, 0x00}, // 0.0
		}

		for {
			line, literals, err := s.liner.Read(func() error { return response.Continuation().Send(s) })
			if err != nil {
				return
			}

			s.logIncoming(string(line), xslices.Map(maps.Keys(literals), func(k string) string {
				return fmt.Sprintf("%v: '%s'", k, literals[k])
			})...)

			// check if we are receiving raw TLS requests, if so skip.
			for _, tlsHeader := range tlsHeaders {
				if bytes.HasPrefix(line, tlsHeader) {
					logrus.Errorf("TLS Handshake detected while not running with TLS/SSL")
					return
				}
			}

			// If the input is not valid UTF-8, we drop the connection.
			if !utf8.Valid(line) {
				reporter.MessageWithContext(ctx,
					"Received invalid UTF-8 command",
					reporter.Context{"line (base 64)": base64.StdEncoding.EncodeToString(line)},
				)

				return
			}

			tag, cmd, err := parse(line, literals, del)
			if err != nil {
				reporter.MessageWithContext(ctx,
					"Failed to parse IMAP command",
					reporter.Context{"error": err},
				)
			} else if cmd.GetStartTLS() != nil {
				// TLS needs to be handled here to ensure that next command read is over the TLS connection.
				if err = s.handleStartTLS(tag, cmd.GetStartTLS()); err != nil {
					logrus.WithError(err).Error("Cannot upgrade connection")
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
