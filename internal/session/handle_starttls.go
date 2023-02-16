package session

import (
	"bufio"
	"crypto/tls"

	"github.com/ProtonMail/gluon/imap/command"
	"github.com/ProtonMail/gluon/internal/response"
)

func (s *Session) handleStartTLS(tag string, _ *command.StartTLS) error {
	if s.tlsConfig == nil {
		return response.No(tag).WithError(ErrTLSUnavailable)
	}

	if err := response.Ok(tag).WithMessage("Begin TLS negotiation now").Send(s); err != nil {
		return err
	}

	conn := tls.Server(s.conn, s.tlsConfig)

	if err := conn.Handshake(); err != nil {
		return err
	}

	s.conn = conn

	s.inputCollector.Reset()
	s.inputCollector.SetSource(bufio.NewReader(s.conn))

	return nil
}
