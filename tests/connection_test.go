package tests

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/ProtonMail/gluon/liner"
	"github.com/bradenaw/juniper/xslices"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
)

func withTag(fn func(string)) {
	fn(uuid.NewString())
}

func lines(lines ...string) string {
	return strings.Join(lines, "\r\n")
}

func repeat(line string, n int) []string {
	var res []string

	for i := 0; i < n; i++ {
		res = append(res, line)
	}

	return res
}

func seq(begin, end int) string {
	var res string

	for i := begin; i < end; i++ {
		res += strconv.Itoa(i) + " "
	}

	return res + strconv.Itoa(end)
}

type testConnection struct {
	tb    testing.TB
	conn  net.Conn
	liner *liner.Liner
}

func newTestConnection(tb testing.TB, conn net.Conn) *testConnection {
	return &testConnection{
		tb:    tb,
		conn:  conn,
		liner: liner.New(conn),
	}
}

func (s *testConnection) C(value string) *testConnection {
	n, err := s.conn.Write([]byte(value + "\r\n"))
	require.NoError(s.tb, err)
	require.Greater(s.tb, n, 0)

	return s
}

func (s *testConnection) Cb(b []byte) *testConnection {
	n, err := s.conn.Write(append(b, []byte("\r\n")...))
	require.NoError(s.tb, err)
	require.Greater(s.tb, n, 0)

	return s
}

func (s *testConnection) Cf(format string, a ...any) *testConnection {
	return s.C(fmt.Sprintf(format, a...))
}

// S expects that the server returns the given lines (in any order).
func (s *testConnection) S(want ...string) *testConnection {
	return s.Sx(xslices.Map(want, func(want string) string { return "^" + regexp.QuoteMeta(want) + "\r\n" })...)
}

// Sx expects that the server returns lines matching the given regexps (in any order).
func (s *testConnection) Sx(want ...string) *testConnection {
	var bad []string

	for _, have := range s.readN(len(want)) {
		if idx := xslices.IndexFunc(want, func(want string) bool {
			return regexp.MustCompile(want).Match(have)
		}); idx >= 0 {
			want = slices.Delete(want, idx, idx+1)
		} else {
			bad = append(bad, string(have))
		}
	}

	if len(bad) > 0 {
		require.Failf(s.tb,
			"Received unexpected responses",
			"want: %q\nbut have:%q",
			want, bad,
		)
	}

	return s
}

// Se expects that the server eventually returns the given lines (in any order).
func (s *testConnection) Se(want ...string) *testConnection {
	return s.Sxe(xslices.Map(want, func(want string) string { return "^" + regexp.QuoteMeta(want) + "\r\n" })...)
}

// Sxe expects that the server eventually returns lines matching the given regexps (in any order).
func (s *testConnection) Sxe(want ...string) *testConnection {
	for len(want) > 0 {
		have := s.read()

		if idx := xslices.IndexFunc(want, func(want string) bool {
			return regexp.MustCompile(want).Match(have)
		}); idx >= 0 {
			want = slices.Delete(want, idx, idx+1)
		}
	}

	return s
}

// Continue is a shortcut for a server continuation request.
func (s *testConnection) Continue() *testConnection {
	s.Sx("\\+")

	return s
}

// OK is a shortcut that we eventually get a tagged OK response of some kind.
func (s *testConnection) OK(tag string, items ...string) {
	want := tag + " OK"

	if len(items) > 0 {
		want += fmt.Sprintf(" [%v]", strings.Join(items, " "))
	}

	s.Sxe(regexp.QuoteMeta(want))
}

// NO is a shortcut that we eventually get a tagged NO response of some kind.
func (s *testConnection) NO(tag string, items ...string) {
	want := tag + " NO"

	if len(items) > 0 {
		want += fmt.Sprintf(" [%v]", strings.Join(items, " "))
	}

	s.Sxe(regexp.QuoteMeta(want))
}

// BAD is a shortcut that we eventually get a tagged BAD response of some kind.
func (s *testConnection) BAD(tag string) {
	s.Sxe(tag + " BAD")
}

// Login is a shortcut for a login request.
func (s *testConnection) Login(username, password string) *testConnection {
	withTag(func(tag string) {
		s.Cf("%v login %v %s", tag, username, password).OK(tag)
	})

	return s
}

func (s *testConnection) doCreateTempDir() (string, func()) {
	name := uuid.NewString()

	// Delete it if it exists already, ignoring the response (OK/NO).
	withTag(func(tag string) {
		s.Cf(`%v DELETE %v`, tag, name).Sx("")
	})

	// Create it (again).
	withTag(func(tag string) {
		s.Cf(`%v CREATE %v`, tag, name).OK(tag)
	})

	// Delete it afterwards.
	return name, func() {
		withTag(func(tag string) {
			s.Cf("%v UNSELECT", tag)
			s.Cf(`%v DELETE %v`, tag, name).OK(tag)
		})
	}
}

func (s *testConnection) doBench(b *testing.B, cmd string) {
	b.Run(cmd, func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			withTag(func(tag string) { s.Cf(`%v %v`, tag, cmd).OK(tag) })
		}
	})
}

// TODO: This is shitty because the uuid of one literal may appear within the data of another literal.
func (s *testConnection) read() []byte {
	line, literals, err := s.liner.Read(func() error { return nil })
	require.NoError(s.tb, err)

	for uuid, literal := range literals {
		line = bytes.Replace(line, []byte(uuid), literal, 1)
	}

	return line
}

func (s *testConnection) readN(n int) [][]byte {
	var res [][]byte

	for i := 0; i < n; i++ {
		res = append(res, s.read())
	}

	return res
}

func (s *testConnection) disconnect() error {
	return s.conn.Close()
}

func (s *testConnection) expectClosed() {
	_, _, err := s.liner.Read(func() error { return nil })
	require.ErrorIs(s.tb, err, io.EOF)
}

func (s *testConnection) upgradeConnection() {
	cert, err := x509.ParseCertificate(testCert.Certificate[0])
	require.NoError(s.tb, err)

	pool := x509.NewCertPool()
	pool.AddCert(cert)

	conn := tls.Client(s.conn, &tls.Config{ServerName: cert.DNSNames[0], RootCAs: pool, MinVersion: tls.VersionTLS13})
	require.NoError(s.tb, conn.Handshake())

	s.conn = conn
	s.liner = liner.New(conn)
}
