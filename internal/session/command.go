package session

import (
	"bytes"
	"context"
	"errors"
	"github.com/ProtonMail/gluon/imap/command"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/logging"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/gluon/rfcparser"
	"github.com/sirupsen/logrus"
)

type commandResult struct {
	command command.Command
	err     error
}

func (s *Session) startCommandReader(ctx context.Context) <-chan commandResult {
	cmdCh := make(chan commandResult)

	logging.GoAnnotated(ctx, func(ctx context.Context) {
		defer close(cmdCh)

		tlsHeaders := [][]byte{
			{0x16, 0x03, 0x01}, // 1.0
			{0x16, 0x03, 0x02}, // 1.1
			{0x16, 0x03, 0x03}, // 1.2
			{0x16, 0x03, 0x04}, // 1.3
			{0x16, 0x00, 0x00}, // 0.0
		}

		parser := command.NewParserWithLiteralContinuationCb(s.scanner, func() error { return response.Continuation().Send(s) })

		for {
			s.inputCollector.Reset()

			cmd, err := parser.Parse()
			s.logIncoming(string(s.inputCollector.Bytes()))
			if err != nil {
				var parserError *rfcparser.Error
				if !errors.As(err, &parserError) {
					return
				}

				if parserError.IsEOF() {
					return
				}

				if err := parser.ConsumeInvalidInput(); err != nil {
					return
				}

				bytesRead := s.inputCollector.Bytes()
				// check if we are receiving raw TLS requests, if so skip.
				for _, tlsHeader := range tlsHeaders {
					if bytes.HasPrefix(bytesRead, tlsHeader) {
						logrus.Errorf("TLS Handshake detected while not running with TLS/SSL")
						return
					}
				}

				logrus.WithError(err).WithField("type", parser.LastParsedCommand()).Error("Failed to parse IMAP command")

				reporter.MessageWithContext(ctx,
					"Failed to parse IMAP command",
					reporter.Context{"error": err, "cmd": parser.LastParsedCommand()},
				)
			} else {
				logrus.Debug(cmd.SanitizedString())
			}

			switch c := cmd.Payload.(type) {
			case *command.StartTLS:
				// TLS needs to be handled here to ensure that next command read is over the TLS connection.
				if err = s.handleStartTLS(cmd.Tag, c); err != nil {
					logrus.WithError(err).Error("Cannot upgrade connection")
					return
				} else {
					continue
				}
			}

			select {
			case cmdCh <- commandResult{command: cmd, err: err}:
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
