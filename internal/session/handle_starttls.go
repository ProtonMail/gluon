package session

import (
	"context"
	"crypto/tls"

	"github.com/ProtonMail/gluon/internal/liner"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
)

func (s *Session) handleStartTLS(ctx context.Context, tag string, cmd *proto.StartTLS) error {
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
	s.liner = liner.New(conn)

	return nil
}
