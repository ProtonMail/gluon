package conn

import (
	"crypto/tls"
	"net"
)

type Listener struct {
	net.Listener
	c *tls.Config
}

func NewListener(l net.Listener, c *tls.Config) *Listener {
	return &Listener{Listener: l, c: c}
}

func (l *Listener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	return &Conn{Conn: c, c: l.c, first: true}, nil
}

type Conn struct {
	net.Conn

	c *tls.Config

	first bool
}

// Read implements the net.Conn interface.
func (c *Conn) Read(b []byte) (int, error) {
	if !c.first {
		return c.Conn.Read(b)
	} else {
		c.first = false
	}

	n, err := c.Conn.Read(b[:1])
	if err != nil {
		return n, err
	}

	if b[0] == 0x16 {
		c.Conn = tls.Server(&pfxConn{Conn: c.Conn, pfx: b[:n]}, c.c)
	}

	return n, nil
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
