package session

import (
	"io"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_pfxConn(t *testing.T) {
	l, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	go func() {
		conn, err := l.Accept()
		require.NoError(t, err)

		if _, err := conn.Write([]byte("world")); err != nil {
			panic(err)
		}

		require.NoError(t, conn.Close())
	}()

	conn, err := net.Dial("tcp", l.Addr().String())
	require.NoError(t, err)

	conn = &pfxConn{
		Conn: conn,
		pfx:  []byte("hello"),
	}

	b, err := io.ReadAll(conn)
	require.NoError(t, err)
	require.Equal(t, "helloworld", string(b))
}
