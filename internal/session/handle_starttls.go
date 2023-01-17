package session

import (
	"crypto/tls"
	"net"

	"github.com/ProtonMail/gluon/internal/response"
)

func (s *Session) handleStartTLS(tag string) error {
	if s.tlsConfig == nil {
		return response.No(tag).WithError(ErrTLSUnavailable)
	}

	if err := response.Ok(tag).WithMessage("Begin TLS negotiation now").Send(s); err != nil {
		return err
	}

	return s.handleUpgrade(nil)
}

func (s *Session) handleUpgrade(line []byte) error {
	// Create a TLS server over the existing connection (replaying the line we just read).
	conn := tls.Server(&pfxConn{Conn: s.conn, pfx: line}, s.tlsConfig)

	// Run the TLS handshake.
	if err := conn.Handshake(); err != nil {
		return err
	}

	// Replace the existing connection with the TLS connection.
	s.conn = conn

	// Reset the liner to use the new connection.
	s.liner.Reset(conn)

	return nil
}

type pfxConn struct {
	net.Conn

	pfx []byte
}

func (c *pfxConn) Read(b []byte) (int, error) {
	if len(c.pfx) == 0 {
		return c.Conn.Read(b)
	}

	n := copy(b, c.pfx)

	c.pfx = c.pfx[n:]

	return n, nil
}
