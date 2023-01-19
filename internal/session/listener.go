package session

import (
	"crypto/tls"
	"net"
)

// MultiConn is a net.Conn that automatically upgrades to TLS when the first byte is read.
type MultiConn struct {
	net.Conn

	tlsConfig *tls.Config
}

// NewMultiConn creates a new MultiConn.
func NewMultiConn(conn net.Conn, tlsConfig *tls.Config) *MultiConn {
	return &MultiConn{
		Conn:      conn,
		tlsConfig: tlsConfig,
	}
}

// Read reads data from the connection.
func (c *MultiConn) Read(b []byte) (int, error) {
	// If the connection is already TLS, just read from it.
	if _, ok := c.Conn.(*tls.Conn); ok {
		return c.Conn.Read(b)
	}

	// Read the first byte.
	n, err := c.Conn.Read(b[:1])
	if err != nil {
		return n, err
	}

	// If the first byte is not 0x16, it's not a TLS handshake.
	if b[0] != 0x16 {
		return n, nil
	}

	// Create a TLS server over the existing connection (replaying the line we just read).
	conn := tls.Server(&pfxConn{Conn: c.Conn, pfx: b[:1]}, c.tlsConfig)

	// Run the TLS handshake.
	if err := conn.Handshake(); err != nil {
		return n, err
	}

	// Replace the existing connection with the TLS connection.
	c.Conn = conn

	return n, nil
}
